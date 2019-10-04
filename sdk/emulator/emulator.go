package emulator

import (
	"sync"
	"time"

	"github.com/dapperlabs/flow-go/pkg/crypto"
	"github.com/dapperlabs/flow-go/pkg/language/runtime"
	"github.com/dapperlabs/flow-go/pkg/types"
	"github.com/dapperlabs/flow-go/sdk/emulator/execution"
	"github.com/dapperlabs/flow-go/sdk/emulator/state"
	etypes "github.com/dapperlabs/flow-go/sdk/emulator/types"
)

// EmulatedBlockchain simulates a blockchain in the background to enable easy smart contract testing.
//
// Contains a versioned World State store and a pending transaction pool for granular state update tests.
// Both "committed" and "intermediate" world states are logged for query by version (state hash) utilities,
// but only "committed" world states are enabled for the SeekToState feature.
type EmulatedBlockchain struct {
	// mapping of committed world states (updated after CommitBlock)
	worldStates map[string][]byte
	// mapping of intermediate world states (updated after SubmitTransaction)
	intermediateWorldStates map[string][]byte
	// current world state
	pendingWorldState *state.WorldState
	// pool of pending transactions waiting to be committed (already executed)
	txPool             map[string]*types.Transaction
	mutex              sync.RWMutex
	computer           *execution.Computer
	rootAccount        types.Account
	rootAccountKey     crypto.PrKey
	lastCreatedAccount types.Account
}

// EmulatedBlockchainOptions is a set of configuration options for an emulated blockchain.
type EmulatedBlockchainOptions struct {
	RootAccountKey crypto.PrKey
	RuntimeLogger  func(string)
}

// DefaultOptions is the default configuration for an emulated blockchain.
var DefaultOptions = &EmulatedBlockchainOptions{
	RuntimeLogger: func(string) {},
}

// NewEmulatedBlockchain instantiates a new blockchain backend for testing purposes.
func NewEmulatedBlockchain(opt *EmulatedBlockchainOptions) *EmulatedBlockchain {
	worldStates := make(map[string][]byte)
	intermediateWorldStates := make(map[string][]byte)
	txPool := make(map[string]*types.Transaction)
	ws := state.NewWorldState()

	worldStates[string(ws.Hash())] = ws.Encode()

	b := &EmulatedBlockchain{
		worldStates:             worldStates,
		intermediateWorldStates: intermediateWorldStates,
		pendingWorldState:       ws,
		txPool:                  txPool,
	}

	runtime := runtime.NewInterpreterRuntime()
	computer := execution.NewComputer(
		runtime,
		opt.RuntimeLogger,
		b.onAccountCreated,
	)

	b.computer = computer
	b.rootAccount, b.rootAccountKey = createRootAccount(ws, opt.RootAccountKey)
	b.lastCreatedAccount = b.rootAccount

	return b
}

func (b *EmulatedBlockchain) RootAccount() types.Address {
	return b.rootAccount.Address
}

func (b *EmulatedBlockchain) RootKey() crypto.PrKey {
	return b.rootAccountKey
}

// GetLatestBlock gets the latest sealed block.
func (b *EmulatedBlockchain) GetLatestBlock() *etypes.Block {
	return b.pendingWorldState.GetLatestBlock()
}

// GetBlockByHash gets a block by hash.
func (b *EmulatedBlockchain) GetBlockByHash(hash crypto.Hash) (*etypes.Block, error) {
	block := b.pendingWorldState.GetBlockByHash(hash)
	if block == nil {
		return nil, &ErrBlockNotFound{BlockHash: hash}
	}

	return block, nil
}

// GetBlockByNumber gets a block by number.
func (b *EmulatedBlockchain) GetBlockByNumber(number uint64) (*etypes.Block, error) {
	block := b.pendingWorldState.GetBlockByNumber(number)
	if block == nil {
		return nil, &ErrBlockNotFound{BlockNum: number}
	}

	return block, nil
}

// GetTransaction gets an existing transaction by hash.
//
// First looks in pending txPool, then looks in current blockchain state.
func (b *EmulatedBlockchain) GetTransaction(txHash crypto.Hash) (*types.Transaction, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if tx, ok := b.txPool[string(txHash)]; ok {
		return tx, nil
	}

	tx := b.pendingWorldState.GetTransaction(txHash)
	if tx == nil {
		return nil, &ErrTransactionNotFound{TxHash: txHash}
	}

	return tx, nil
}

// GetTransactionAtVersion gets an existing transaction by hash at a specified state.
func (b *EmulatedBlockchain) GetTransactionAtVersion(txHash, version crypto.Hash) (*types.Transaction, error) {
	ws, err := b.getWorldStateAtVersion(version)
	if err != nil {
		return nil, err
	}

	tx := ws.GetTransaction(txHash)
	if tx == nil {
		return nil, &ErrTransactionNotFound{TxHash: txHash}
	}

	return tx, nil
}

// GetAccount gets account information associated with an address identifier.
func (b *EmulatedBlockchain) GetAccount(address types.Address) (*types.Account, error) {
	registers := b.pendingWorldState.Registers.NewView()
	runtimeContext := execution.NewRuntimeContext(registers)
	account := runtimeContext.GetAccount(address)
	if account == nil {
		return nil, &ErrAccountNotFound{Address: address}
	}

	return account, nil
}

// GetAccountAtVersion gets account information associated with an address identifier at a specified state.
func (b *EmulatedBlockchain) GetAccountAtVersion(address types.Address, version crypto.Hash) (*types.Account, error) {
	ws, err := b.getWorldStateAtVersion(version)
	if err != nil {
		return nil, err
	}

	registers := ws.Registers.NewView()
	runtimeContext := execution.NewRuntimeContext(registers)
	account := runtimeContext.GetAccount(address)
	if account == nil {
		return nil, &ErrAccountNotFound{Address: address}
	}

	return account, nil
}

// SubmitTransaction sends a transaction to the network that is immediately executed (updates blockchain state).
//
// Note that the resulting state is not finalized until CommitBlock() is called.
// However, the pending blockchain state is indexed for testing purposes.
func (b *EmulatedBlockchain) SubmitTransaction(tx *types.Transaction) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.txPool[string(tx.Hash())]; exists {
		return &ErrDuplicateTransaction{TxHash: tx.Hash()}
	}

	if b.pendingWorldState.ContainsTransaction(tx.Hash()) {
		return &ErrDuplicateTransaction{TxHash: tx.Hash()}
	}

	if err := b.verifySignatures(tx); err != nil {
		return err
	}

	b.txPool[string(tx.Hash())] = tx
	b.pendingWorldState.InsertTransaction(tx)

	registers := b.pendingWorldState.Registers.NewView()

	err := b.computer.ExecuteTransaction(registers, tx)
	if err != nil {
		b.pendingWorldState.UpdateTransactionStatus(tx.Hash(), types.TransactionReverted)

		b.updatePendingWorldStates(tx.Hash())

		return &ErrTransactionReverted{TxHash: tx.Hash(), Err: err}
	}

	b.pendingWorldState.SetRegisters(registers.UpdatedRegisters())
	b.pendingWorldState.UpdateTransactionStatus(tx.Hash(), types.TransactionFinalized)

	b.updatePendingWorldStates(tx.Hash())

	return nil
}

func (b *EmulatedBlockchain) updatePendingWorldStates(txHash crypto.Hash) {
	if _, exists := b.intermediateWorldStates[string(txHash)]; exists {
		return
	}

	bytes := b.pendingWorldState.Encode()
	b.intermediateWorldStates[string(txHash)] = bytes
}

// CallScript executes a read-only script against the world state and returns the result.
func (b *EmulatedBlockchain) CallScript(script []byte) (interface{}, error) {
	registers := b.pendingWorldState.Registers.NewView()
	return b.computer.ExecuteScript(registers, script)
}

// CallScriptAtVersion executes a read-only script against a specified world state and returns the result.
func (b *EmulatedBlockchain) CallScriptAtVersion(script []byte, version crypto.Hash) (interface{}, error) {
	ws, err := b.getWorldStateAtVersion(version)
	if err != nil {
		return nil, err
	}

	registers := ws.Registers.NewView()
	return b.computer.ExecuteScript(registers, script)
}

// CommitBlock takes all pending transactions and commits them into a block.
//
// Note that this clears the pending transaction pool and indexes the committed
// blockchain state for testing purposes.
func (b *EmulatedBlockchain) CommitBlock() *etypes.Block {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	txHashes := make([]crypto.Hash, 0)
	for _, tx := range b.txPool {
		txHashes = append(txHashes, tx.Hash())
		if b.pendingWorldState.GetTransaction(tx.Hash()).Status != types.TransactionReverted {
			b.pendingWorldState.UpdateTransactionStatus(tx.Hash(), types.TransactionSealed)
		}
	}
	b.txPool = make(map[string]*types.Transaction)

	prevBlock := b.pendingWorldState.GetLatestBlock()
	block := &etypes.Block{
		Number:            prevBlock.Number + 1,
		Timestamp:         time.Now(),
		PreviousBlockHash: prevBlock.Hash(),
		TransactionHashes: txHashes,
	}

	b.pendingWorldState.InsertBlock(block)
	b.commitWorldState(block.Hash())

	return block
}

// SeekToState rewinds the blockchain state to a previously committed history.
//
// Note that this only seeks to a committed world state (not intermediate world state)
// and this clears all pending transactions in txPool.
func (b *EmulatedBlockchain) SeekToState(hash crypto.Hash) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if bytes, ok := b.worldStates[string(hash)]; ok {
		ws := state.Decode(bytes)
		b.pendingWorldState = ws
		b.txPool = make(map[string]*types.Transaction)
	}
}

func (b *EmulatedBlockchain) getWorldStateAtVersion(wsHash crypto.Hash) (*state.WorldState, error) {
	if wsBytes, ok := b.worldStates[string(wsHash)]; ok {
		return state.Decode(wsBytes), nil
	}

	if wsBytes, ok := b.intermediateWorldStates[string(wsHash)]; ok {
		return state.Decode(wsBytes), nil
	}

	return nil, &ErrInvalidStateVersion{Version: wsHash}
}

func (b *EmulatedBlockchain) commitWorldState(blockHash crypto.Hash) {
	if _, exists := b.worldStates[string(blockHash)]; exists {
		return
	}

	bytes := b.pendingWorldState.Encode()
	b.worldStates[string(blockHash)] = bytes
}

func (b *EmulatedBlockchain) onAccountCreated(account types.Account) {
	b.lastCreatedAccount = account
}

// lastCreatedAccount returns the last account that was created in the blockchain.
func (b *EmulatedBlockchain) LastCreatedAccount() types.Account {
	return b.lastCreatedAccount
}

// verifySignatures verifies that a transaction contains the necessary signatures.
//
// An error is returned if any of the expected signatures are invalid or missing.
func (b *EmulatedBlockchain) verifySignatures(tx *types.Transaction) error {
	accountWeights := make(map[types.Address]int)

	for _, accountSig := range tx.Signatures {
		err := b.verifyAccountSignature(accountSig, tx.CanonicalEncoding())
		if err != nil {
			return err
		}

		// TODO: add key weights
		accountWeights[accountSig.Account] = 1
	}

	if accountWeights[tx.PayerAccount] < 1 {
		return &ErrMissingSignature{tx.PayerAccount}
	}

	for _, account := range tx.ScriptAccounts {
		if accountWeights[account] < 1 {
			return &ErrMissingSignature{account}
		}
	}

	return nil
}

// verifyAccountSignature verifies a that an account signature is valid for the account and given message.
//
// An error is returned if the account does not contain a public key that correctly verifies the signature
// against the given message.
func (b *EmulatedBlockchain) verifyAccountSignature(
	accountSig types.AccountSignature,
	message []byte,
) error {
	account, err := b.GetAccount(accountSig.Account)
	if err != nil {
		return &ErrInvalidSignatureAccount{Account: accountSig.Account}
	}

	signature := crypto.Signature(accountSig.Signature)

	// TODO: replace hard-coded signature algorithm
	salg, _ := crypto.NewSignatureAlgo(crypto.ECDSA_P256)

	// TODO: account signatures should specify a public key (possibly by index) to avoid this loop
	for _, publicKeyBytes := range account.PublicKeys {
		publicKey, err := salg.DecodePubKey(publicKeyBytes)
		if err != nil {
			continue
		}

		// TODO: replace hard-coded hashing algorithm
		hasher, _ := crypto.NewHashAlgo(crypto.SHA3_256)

		valid, err := salg.VerifyBytes(publicKey, signature, message, hasher)
		if err != nil {
			continue
		}

		if valid {
			return nil
		}
	}

	return &ErrInvalidSignaturePublicKey{
		Account: accountSig.Account,
	}
}

// createRootAccount creates a new root account and commits it to the world state.
func createRootAccount(ws *state.WorldState, prKey crypto.PrKey) (types.Account, crypto.PrKey) {
	registers := ws.Registers.NewView()

	// TODO: replace hard-coded signature algorithm
	salg, _ := crypto.NewSignatureAlgo(crypto.ECDSA_P256)

	if prKey == nil {
		prKey, _ = salg.GeneratePrKey([]byte("elephant ears"))
	}

	pubKeyBytes, _ := salg.EncodePubKey(prKey.Pubkey())

	runtimeContext := execution.NewRuntimeContext(registers)
	accountID, _ := runtimeContext.CreateAccount(pubKeyBytes, []byte{})

	ws.SetRegisters(registers.UpdatedRegisters())

	accountAddress := types.BytesToAddress(accountID)
	account := runtimeContext.GetAccount(accountAddress)

	return *account, prKey
}
