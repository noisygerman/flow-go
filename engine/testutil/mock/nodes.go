package mock

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog"

	collectioningest "github.com/dapperlabs/flow-go/engine/collection/ingest"
	"github.com/dapperlabs/flow-go/engine/collection/provider"
	consensusingest "github.com/dapperlabs/flow-go/engine/consensus/ingestion"
	"github.com/dapperlabs/flow-go/engine/consensus/matching"
	"github.com/dapperlabs/flow-go/engine/consensus/propagation"
	"github.com/dapperlabs/flow-go/engine/execution/execution"
	"github.com/dapperlabs/flow-go/engine/execution/execution/state"
	"github.com/dapperlabs/flow-go/engine/execution/execution/virtualmachine"
	"github.com/dapperlabs/flow-go/engine/execution/ingestion"
	"github.com/dapperlabs/flow-go/engine/execution/receipts"
	"github.com/dapperlabs/flow-go/module"
	"github.com/dapperlabs/flow-go/module/mempool"
	"github.com/dapperlabs/flow-go/module/trace"
	"github.com/dapperlabs/flow-go/network"
	"github.com/dapperlabs/flow-go/network/stub"
	"github.com/dapperlabs/flow-go/protocol"
	"github.com/dapperlabs/flow-go/storage"
	"github.com/dapperlabs/flow-go/storage/ledger/databases/leveldb"
)

// GenericNode implements a generic in-process node for tests.
type GenericNode struct {
	Log    zerolog.Logger
	Tracer trace.Tracer
	DB     *badger.DB
	State  protocol.State
	Me     module.Local
	Net    *stub.Network
}

// CollectionNode implements an in-process collection node for tests.
type CollectionNode struct {
	GenericNode
	Pool            mempool.Transactions
	Collections     storage.Collections
	IngestionEngine *collectioningest.Engine
	ProviderEngine  *provider.Engine
}

// ConsensusNode implements an in-process consensus node for tests.
type ConsensusNode struct {
	GenericNode
	Guarantees        mempool.Guarantees
	Approvals         mempool.Approvals
	Receipts          mempool.Receipts
	Seals             mempool.Seals
	IngestionEngine   *consensusingest.Engine
	PropagationEngine *propagation.Engine
	MatchingEngine    *matching.Engine
}

// ExecutionNode implements a mocked execution node for tests.
type ExecutionNode struct {
	GenericNode
	BlocksEngine    *ingestion.Engine
	ExecutionEngine *execution.Engine
	ReceiptsEngine  *receipts.Engine
	BadgerDB        *badger.DB
	LevelDB         *leveldb.LevelDB
	VM              virtualmachine.VirtualMachine
	State           state.ExecutionState
}

func (en ExecutionNode) Done() {
	<-en.BlocksEngine.Done()
	<-en.ExecutionEngine.Done()
	<-en.ReceiptsEngine.Done()
	en.BadgerDB.Close()
	en.LevelDB.SafeClose()
}

// VerificationNode implements an in-process verification node for tests.
type VerificationNode struct {
	GenericNode
	Receipts       mempool.Receipts
	Blocks         mempool.Blocks
	Collections    mempool.Collections
	ChunkStates    mempool.ChunkStates
	IngestEngine   network.Engine
	VerifierEngine network.Engine
}
