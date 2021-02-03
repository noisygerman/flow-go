package assigner

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/onflow/flow-go/engine"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/state/protocol"
	"github.com/onflow/flow-go/storage"
	"github.com/onflow/flow-go/utils/logging"
)

type Engine struct {
	unit    *engine.Unit
	log     zerolog.Logger
	metrics module.VerificationMetrics
	me      module.Local
	state   protocol.State

	headerStorage    storage.Headers       // used to check block existence before verifying
	assigner         module.ChunkAssigner  // used to determine chunks this node needs to verify
	chunksQueue      storage.ChunksQueue   // to store chunks to be verified
	newChunkListener module.NewJobListener // to notify about a new chunk
	finishProcessing finishProcessing      // to report a block has been processed
}

func New(
	log zerolog.Logger,
	metrics module.VerificationMetrics,
	tracer module.Tracer,
	me module.Local,
	state protocol.State,
	headerStorage storage.Headers,
	assigner module.ChunkAssigner,
	chunksQueue storage.ChunksQueue,
	newChunkListener module.NewJobListener,
) *Engine {
	e := &Engine{
		unit:             engine.NewUnit(),
		log:              log.With().Str("engine", "assigner").Logger(),
		metrics:          metrics,
		me:               me,
		state:            state,
		headerStorage:    headerStorage,
		assigner:         assigner,
		chunksQueue:      chunksQueue,
		newChunkListener: newChunkListener,
	}

	return e
}

func (e *Engine) withFinishProcessing(finishProcessing finishProcessing) {
	e.finishProcessing = finishProcessing
}

func (e *Engine) Ready() <-chan struct{} {
	return e.unit.Ready()
}

func (e *Engine) Done() <-chan struct{} {
	return e.unit.Done()
}

func (e *Engine) SubmitLocal(event interface{}) {
	e.Submit(e.me.NodeID(), event)
}

func (e *Engine) Submit(originID flow.Identifier, event interface{}) {
	e.unit.Launch(func() {
		err := e.Process(originID, event)
		if err != nil {
			engine.LogError(e.log, err)
		}
	})
}

func (e *Engine) ProcessLocal(event interface{}) error {
	return e.Process(e.me.NodeID(), event)
}

func (e *Engine) Process(originID flow.Identifier, event interface{}) error {
	return e.unit.Do(func() error {
		return e.process(originID, event)
	})
}

func (e *Engine) process(originID flow.Identifier, event interface{}) error {
	return nil
}

func (e *Engine) ProcessFinalizedBlock(block *flow.Block) {
	blockID := block.ID()
	e.log.Debug().
		Hex("block_id", logging.ID(blockID)).
		Msg("new finalized block received")

	for _, receipt := range block.Payload.Receipts {
		e.handleExecutionReceipt(receipt, blockID)
	}
}

// handleExecutionReceipt
func (e *Engine) handleExecutionReceipt(receipt *flow.ExecutionReceipt, containerBlockID flow.Identifier) {
	receiptID := receipt.ID()
	resultID := receipt.ExecutionResult.ID()
	referenceBlockID := receipt.ExecutionResult.BlockID

	log := e.log.With().
		Hex("container_block_id", logging.ID(containerBlockID)).
		Hex("reference_block_id", logging.ID(referenceBlockID)).
		Hex("receipt_id", logging.ID(receiptID)).
		Hex("result_id", logging.ID(resultID)).Logger()

	// monitoring: increases number of received execution receipts
	e.metrics.OnExecutionReceiptReceived()

	// verification node should be staked at reference block id
	staked, err := e.stakedAtBlockID(referenceBlockID)
	if err != nil {
		log.Debug().Err(err).Msg("could not verify stake of verification node for result")
	}

	if !staked {
		log.Debug().Msg("discards result for unstaked reference block")
	}

}

// stakedAtBlockID checks whether this instance of node has staked at specified block ID as a verification node.
// It returns true and nil if verification node has staked at specified block ID, and returns false, and nil otherwise.
// It returns false and error if it could not extract the stake of (verification node) node at the specified block.
func (e *Engine) stakedAtBlockID(blockID flow.Identifier) (bool, error) {
	// extracts identity of verification node at block height of result
	myIdentity, err := protocol.StakedIdentity(e.state, blockID, e.me.NodeID())
	if err != nil {
		return false, fmt.Errorf("could not check if node is staked at block %v: %w", blockID, err)
	}

	if myIdentity.Role != flow.RoleVerification {
		return false, nil
	}
	return true, nil
}
