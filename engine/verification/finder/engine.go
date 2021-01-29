package finder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog"

	"github.com/onflow/flow-go/consensus/hotstuff/model"
	"github.com/onflow/flow-go/engine"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/model/verification"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/mempool"
	"github.com/onflow/flow-go/module/trace"
	"github.com/onflow/flow-go/network"
	"github.com/onflow/flow-go/state/protocol"
	"github.com/onflow/flow-go/storage"
	"github.com/onflow/flow-go/utils/logging"
)

// Engine receives receipts and passes them to the match engine if the block of the receipt is available
// and the verification node is staked at the block ID of the result part of receipt.
//
// A receipt follows a lifecycle in this engine:
// cached: the receipt is received but not handled yet.
// pending: the receipt is handled, but its corresponding block has not received at this node yet.
// discarded: the receipt's block has received, but this verification node has not staked at block of the receipt.
// ready: the receipt's block has received, and this verification node is staked for that block,
// hence receipt's result is  ready to be forwarded to match engine
// processed: the receipt's result has been forwarded to matching engine.
//
// This engine ensures that each (ready) result is passed to match engine only once.
// Hence, among concurrent ready receipts with shared result, only one instance of result is passed to match engine.
type Engine struct {
	unit               *engine.Unit
	log                zerolog.Logger
	metrics            module.VerificationMetrics
	me                 module.Local
	match              network.Engine
	state              protocol.State
	readyReceipts      mempool.ReceiptDataPacks // used to keep the receipts ready for process
	blocks             storage.Blocks           // used to extract receipts from finalized blocks
	headerStorage      storage.Headers          // used to check block existence before verifying
	processedResultIDs mempool.Identifiers      // used to keep track of the processed results
	discardedResultIDs mempool.Identifiers      // used to keep track of discarded results while node was not staked for epoch
	blockIDsCache      mempool.Identifiers      // used as a cache to keep track of new finalized blocks
	receiptIDsByResult mempool.IdentifierMap    // used as a mapping to keep track of receipts with the same result
	processInterval    time.Duration            // used to define intervals at which engine moves receipts through pipeline
	tracer             module.Tracer
}

func New(
	log zerolog.Logger,
	metrics module.VerificationMetrics,
	tracer module.Tracer,
	net module.Network,
	me module.Local,
	state protocol.State,
	match network.Engine,
	readyReceipts mempool.ReceiptDataPacks,
	headerStorage storage.Headers,
	processedResultIDs mempool.Identifiers,
	discardedResultIDs mempool.Identifiers,
	receiptsIDsByResult mempool.IdentifierMap,
	blockIDsCache mempool.Identifiers,
	processInterval time.Duration,
) (*Engine, error) {
	e := &Engine{
		unit:               engine.NewUnit(),
		log:                log.With().Str("engine", "finder").Logger(),
		metrics:            metrics,
		me:                 me,
		state:              state,
		match:              match,
		headerStorage:      headerStorage,
		readyReceipts:      readyReceipts,
		processedResultIDs: processedResultIDs,
		discardedResultIDs: discardedResultIDs,
		receiptIDsByResult: receiptsIDsByResult,
		blockIDsCache:      blockIDsCache,
		processInterval:    processInterval,
		tracer:             tracer,
	}

	_, err := net.Register(engine.ReceiveReceipts, e)
	if err != nil {
		return nil, fmt.Errorf("could not register engine on execution receipt provider channel: %w", err)
	}
	return e, nil
}

// Ready returns a channel that is closed when the finder engine is ready.
func (e *Engine) Ready() <-chan struct{} {
	// Runs a periodic check to iterate over receipts and move them through the pipeline.
	// If onTimer takes longer than processInterval, the next call will be blocked until the previous
	// call has finished. That being said, there won't be two onTimer running in parallel.
	// See test cases for LaunchPeriodically
	e.unit.LaunchPeriodically(e.onTimer, e.processInterval, time.Duration(0))
	return e.unit.Ready()
}

// Done returns a channel that is closed when the verifier engine is done.
func (e *Engine) Done() <-chan struct{} {
	return e.unit.Done()
}

// SubmitLocal submits an event originating on the local node.
func (e *Engine) SubmitLocal(event interface{}) {
	e.Submit(e.me.NodeID(), event)
}

// Submit submits the given event from the node with the given origin ID
// for processing in a non-blocking manner. It returns instantly and logs
// a potential processing error internally when done.
func (e *Engine) Submit(originID flow.Identifier, event interface{}) {
	e.unit.Launch(func() {
		err := e.Process(originID, event)
		if err != nil {
			engine.LogError(e.log, err)
		}
	})
}

// ProcessLocal processes an event originating on the local node.
func (e *Engine) ProcessLocal(event interface{}) error {
	return e.Process(e.me.NodeID(), event)
}

// Process processes the given event from the node with the given origin ID in
// a blocking manner. The current version of finder engine does not expect receiving any event from
// networking layer. Hence, Process method returns an error upon invocation.
func (e *Engine) Process(originID flow.Identifier, event interface{}) error {
	return e.unit.Do(func() error {
		return fmt.Errorf("finder engine does not expect events through process")
	})
}

// handleExecutionReceipt receives an execution receipt and adds it to the ready receipt mempool.
func (e *Engine) handleExecutionReceipt(receipt *flow.ExecutionReceipt) {
	// initiates tracer for receipt.
	span, ok := e.tracer.GetSpan(receipt.ID(), trace.VERProcessExecutionReceipt)
	ctx := context.Background()
	if !ok {
		span = e.tracer.StartSpan(receipt.ID(), trace.VERProcessExecutionReceipt)
		span.SetTag("execution_receipt_id", receipt.ID())
		defer span.Finish()
	}
	ctx = opentracing.ContextWithSpan(ctx, span)

	e.tracer.WithSpanFromContext(ctx, trace.VERFindHandleExecutionReceipt, func() {
		receiptID := receipt.ID()
		resultID := receipt.ExecutionResult.ID()
		blockID := receipt.ExecutionResult.BlockID

		log := e.log.With().
			Hex("block_id", logging.ID(blockID)).
			Hex("receipt_id", logging.ID(receiptID)).
			Hex("result_id", logging.ID(resultID)).Logger()
		log.Info().
			Msg("execution receipt arrived")

		// monitoring: increases number of received execution receipts
		e.metrics.OnExecutionReceiptReceived()

		// checks if the result has already been processed or discarded
		if e.processedResultIDs.Has(resultID) {
			log.Debug().Msg("drops handling already processed result")
			return
		}
		if e.discardedResultIDs.Has(resultID) {
			log.Debug().Msg("drops handling already discarded result")
			return
		}

		// checks whether verification node is staked at snapshot of this result's block.
		ok, err := e.stakedAtBlockID(blockID)
		if err != nil {
			e.log.Debug().
				Err(err).
				Msg("unable to verify stake of node at block id of receipt")
			return
		}
		if !ok {
			discarded := e.discardedResultIDs.Add(resultID)
			log.Debug().
				Bool("added_to_discard_pool", discarded).
				Msg("execution result marks discarded")
			return
		}

		// adds receipt to ready mempool
		receiptDataPack := &verification.ReceiptDataPack{
			Receipt: receipt,
			Ctx:     ctx,
		}
		added := e.readyReceipts.Add(receiptDataPack)
		e.log.Debug().
			Bool("added", added).
			Msg("adding receipt data pack to ready mempool")

		err = e.receiptIDsByResult.Append(resultID, receiptID)
		if err != nil {
			e.log.Debug().
				Err(err).
				Msg("could not append receipt to receipt-ids-by-result mempool")
		}

		log.Debug().
			Msg("execution receipt successfully handled")
	})
}

// To implement FinalizationConsumer
func (e *Engine) OnBlockIncorporated(*model.Block) {

}

// OnFinalizedBlock is part of implementing FinalizationConsumer interface
// On receiving a block, it caches the block ID to be checked in the next onTimer loop.
//
// OnFinalizedBlock notifications are produced by the Finalization Logic whenever
// a block has been finalized. They are emitted in the order the blocks are finalized.
// Prerequisites:
// Implementation must be concurrency safe; Non-blocking;
// and must handle repetition of the same events (with some processing overhead).
func (e *Engine) OnFinalizedBlock(block *model.Block) {
	ok := e.blockIDsCache.Add(block.BlockID)
	e.log.Debug().
		Bool("added_new_blocks", ok).
		Hex("block_id", logging.ID(block.BlockID)).
		Msg("new finalized block received")
}

// To implement FinalizationConsumer
func (e *Engine) OnDoubleProposeDetected(*model.Block, *model.Block) {}

// stakedAtBlockID checks whether this instance of verification node has staked at specified block ID.
// It returns true an  nil if verification node has staked at specified block ID, and returns false, and nil otherwise.
// It returns false and error if it could not extract the stake of (verification node) node at the specified block.
func (e *Engine) stakedAtBlockID(blockID flow.Identifier) (bool, error) {
	// extracts identity of verification node at block height of result
	staked, err := protocol.IsNodeStakedAtBlockID(e.state, blockID, e.me.NodeID())
	if err != nil {
		return false, fmt.Errorf("could not check if node is staked at block %v: %w", blockID, err)
	}
	return staked, nil
}

// processResult submits the result to the match engine.
// originID is the identifier of the node that initially sends a receipt containing this result.
// It returns true and nil if the result is submitted successfully to the match engine.
// Otherwise, it returns false, and error if the result is not going successfully to the match engine. It returns false,
// and nil, if the result has already been processed.
func (e *Engine) processResult(ctx context.Context, originID flow.Identifier, result *flow.ExecutionResult) (bool, error) {
	span, _ := e.tracer.StartSpanFromContext(ctx, trace.VERFindProcessResult)
	defer span.Finish()

	resultID := result.ID()
	log := e.log.With().Hex("result_id", logging.ID(resultID)).Logger()
	if e.processedResultIDs.Has(resultID) {
		log.Debug().Msg("result already processed")
		return false, nil
	}
	if e.discardedResultIDs.Has(resultID) {
		e.log.Debug().Msg("drops handling already discarded result")
		return false, nil
	}

	err := e.match.Process(originID, result)
	if err != nil {
		return false, fmt.Errorf("submission error to match engine: %w", err)
	}

	log.Info().Msg("result submitted to match engine")

	// monitoring: increases number of execution results sent
	e.metrics.OnExecutionResultSent()

	return true, nil
}

// onResultProcessed is called whenever a result is processed completely and
// is passed to the match engine. It marks the result as processed, and removes
// all receipts with the same result from mempool.
func (e *Engine) onResultProcessed(ctx context.Context, resultID flow.Identifier) {
	e.tracer.WithSpanFromContext(ctx, trace.VERFindOnResultProcessed, func() {
		log := e.log.With().Hex("result_id", logging.ID(resultID)).Logger()
		// marks result as processed
		added := e.processedResultIDs.Add(resultID)
		log.Debug().
			Bool("added_to_processed_mempool", added).
			Msg("marking result as processed")

		// extracts and removes all receipt ids with this result
		receiptIDs, ok := e.receiptIDsByResult.Get(resultID)
		if !ok {
			log.Debug().Msg("could not retrieve receipt ids associated with this result")
		}
		removed := e.receiptIDsByResult.Rem(resultID)
		log.Debug().
			Bool("removed", removed).
			Msg("removing processed result ids")

		// drops all receipts with the same result from ready mempool
		for _, receiptID := range receiptIDs {
			removed := e.readyReceipts.Rem(receiptID)
			log.Debug().
				Bool("removed", removed).
				Hex("receipt_id", logging.ID(receiptID)).
				Msg("cleaning receipts of processed result")
		}
	})
}

// checkCachedBlocks iterates over the new cached finalized blocks, and handles their included execution receipts.
func (e *Engine) checkCachedBlocks() {
	for _, blockID := range e.blockIDsCache.All() {
		// removes block identifier from cache
		removed := e.blockIDsCache.Rem(blockID)
		log := e.log.With().
			Hex("block_id", logging.ID(blockID)).
			Logger()

		log.Debug().
			Bool("removed", removed).
			Msg("removes block id from cached block ids")

		// extracts receipts from block
		receipts, err := e.receipts(blockID)
		if err != nil {
			log.Debug().Err(err).Msg("could not extract receipts from finalized block")
			continue
		}

		log.Debug().
			Int("receipt_num", len(receipts)).
			Msg("receipts retrieved successfully from finalized block")

		for _, receipt := range receipts {
			e.handleExecutionReceipt(receipt)
		}
	}
}

// checkReadyReceipts iterates over receipts ready for process and processes them.
func (e *Engine) checkReadyReceipts() {
	for _, rdp := range e.readyReceipts.All() {
		e.tracer.WithSpanFromContext(rdp.Ctx, trace.VERFindCheckReadyReceipts, func() {
			receiptID := rdp.Receipt.ID()
			resultID := rdp.Receipt.ExecutionResult.ID()

			ok, err := e.processResult(rdp.Ctx, rdp.OriginID, &rdp.Receipt.ExecutionResult)
			if err != nil {
				e.log.Error().
					Err(err).
					Hex("receipt_id", logging.ID(receiptID)).
					Hex("result_id", logging.ID(resultID)).
					Msg("could not process result")
				return
			}

			if !ok {
				// result has already been processed, no cleanup is needed
				return
			}

			// performs clean up
			e.onResultProcessed(rdp.Ctx, resultID)

			e.log.Debug().
				Hex("receipt_id", logging.ID(receiptID)).
				Hex("result_id", logging.ID(resultID)).
				Msg("result processed successfully")
		})

	}
}

// onTimer is called periodically by the unit module of Finder engine.
// It encapsulates the set of handlers should be executed periodically in order.
func (e *Engine) onTimer() {
	wg := &sync.WaitGroup{}

	wg.Add(2)

	go func() {
		e.checkCachedBlocks()
		wg.Done()
	}()

	go func() {
		e.checkReadyReceipts()
		wg.Done()
	}()

	wg.Wait()
}

// receipts extracts and returns all ExecutionReceipts from finalized block.
func (e Engine) receipts(blockID flow.Identifier) ([]*flow.ExecutionReceipt, error) {
	block, err := e.blocks.ByID(blockID)
	if err != nil {
		return nil, fmt.Errorf("could not extract block from storage: %w", err)
	}

	return block.Payload.Receipts, nil
}
