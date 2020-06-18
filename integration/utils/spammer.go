package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	flowsdk "github.com/onflow/flow-go-sdk"
	"google.golang.org/grpc"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/utils/unittest"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/examples"
	"github.com/onflow/flow-go-sdk/templates"
)

const (
	// Pinned to specific commit
	// More transactions listed here: https://github.com/onflow/flow-ft/tree/0e8024a483ce85c06eb165c2d4c9a5795ba167a1/transactions
	FungibleTokenTransactionsBaseURL = "https://raw.githubusercontent.com/onflow/flow-ft/0e8024a483ce85c06eb165c2d4c9a5795ba167a1/src/transactions/"
	TransferTokens                   = "transfer_tokens.cdc"
)

// This should only used for testing reasons
type flowAccount struct {
	address    *flowsdk.Address
	accountKey *flowsdk.AccountKey
	signer     crypto.InMemorySigner
	seqNumber  uint64
	signerLock sync.Mutex
}

func newFlowAccount(address *flowsdk.Address,
	accountKey *flowsdk.AccountKey,
	signer crypto.InMemorySigner) *flowAccount {
	return &flowAccount{address: address,
		accountKey: accountKey,
		signer:     signer,
		seqNumber:  uint64(0),
		signerLock: sync.Mutex{},
	}
}

type LoadGenerator struct {
	numberOfAccounts     int
	flowClient           *client.Client
	serviceAccount       *flowAccount
	flowTokenAddress     *flowsdk.Address
	fungibleTokenAddress *flowsdk.Address
	accounts             []*flowAccount
	step                 int
	txTracker            *TxTracker
}

// In case we have to creat the client
// flowClient, err := client.New(gs.accessAddr, grpc.WithInsecure())
// require.NoError(gs.T(), err, "could not get client")

func decodeAddressFromHex(hexinput string) (*flowsdk.Address, error) {
	var output flowsdk.Address
	inputBytes, err := hex.DecodeString(hexinput)
	if err != nil {
		return nil, err
	}
	copy(output[:], inputBytes)
	return &output, nil
}

// TODO flowsdk.Testnet as chainID
// TODO remove the need for servAccPrivKeyHex when we open it up to everyone
func NewLoadGenerator(fclient *client.Client,
	servAccPrivKeyHex string,
	serviceAccountAddressHex string,
	fungibleTokenAddressHex string,
	flowTokenAddressHex string,
	numberOfAccounts int) (*LoadGenerator, error) {

	serviceAccountAddress, err := decodeAddressFromHex(serviceAccountAddressHex)
	if err != nil {
		return nil, err
	}

	fungibleTokenAddress, err := decodeAddressFromHex(fungibleTokenAddressHex)
	if err != nil {
		return nil, err
	}

	flowTokenAddress, err := decodeAddressFromHex(flowTokenAddressHex)
	if err != nil {
		return nil, err
	}

	servAcc, err := loadServiceAccount(fclient, serviceAccountAddress, servAccPrivKeyHex)
	if err != nil {
		return nil, fmt.Errorf("error loading service account %w", err)
	}

	txTracker, err := NewTxTracker(1000)
	if err != nil {
		return nil, err
	}

	lGen := &LoadGenerator{
		numberOfAccounts:     numberOfAccounts,
		flowClient:           fclient,
		serviceAccount:       servAcc,
		fungibleTokenAddress: fungibleTokenAddress,
		flowTokenAddress:     flowTokenAddress,
		accounts:             make([]*flowAccount, 0),
		step:                 0,
		txTracker:            txTracker,
	}
	return lGen, nil
}

func loadServiceAccount(flowClient *client.Client,
	servAccAddress *flowsdk.Address,
	servAccPrivKeyHex string) (*flowAccount, error) {

	acc, err := flowClient.GetAccount(context.Background(), *servAccAddress)
	if err != nil {
		return nil, fmt.Errorf("error while calling get account for service account %w", err)
	}
	accountKey := acc.Keys[0]

	privateKey, err := crypto.DecodePrivateKeyHex(accountKey.SigAlgo, servAccPrivKeyHex)
	if err != nil {
		return nil, fmt.Errorf("error while decoding serice account private key hex %w", err)
	}

	signer := crypto.NewInMemorySigner(privateKey, accountKey.HashAlgo)

	return &flowAccount{
		address:    servAccAddress,
		accountKey: accountKey,
		seqNumber:  accountKey.SequenceNumber,
		signer:     signer,
		signerLock: sync.Mutex{},
	}, nil
}

func (cg *LoadGenerator) Next() error {

	ref, err := cg.flowClient.GetLatestBlockHeader(context.Background(), false)
	examples.Handle(err)

	// add keys to service account
	if cg.step == 0 {
		publicKeysStr := strings.Builder{}

		for i := 0; i < cg.numberOfAccounts; i++ {
			publicKeysStr.WriteString("signer.addPublicKey(")
			publicKeysStr.WriteString(languageEncodeBytes(cg.serviceAccount.accountKey.Encode()))
			publicKeysStr.WriteString(")\n")
		}
		script := fmt.Sprintf(`
		transaction {
		prepare(signer: AuthAccount) {
				%s
			}
		}`, publicKeysStr.String())

		addKeysTx := flowsdk.NewTransaction().
			SetReferenceBlockID(ref.ID).
			SetScript([]byte(script)).
			SetProposalKey(*cg.serviceAccount.address, cg.serviceAccount.accountKey.ID, cg.serviceAccount.accountKey.SequenceNumber).
			SetPayer(*cg.serviceAccount.address).
			AddAuthorizer(*cg.serviceAccount.address)

		cg.serviceAccount.signerLock.Lock()
		defer cg.serviceAccount.signerLock.Unlock()

		err := addKeysTx.SignEnvelope(*cg.serviceAccount.address, cg.serviceAccount.accountKey.ID, cg.serviceAccount.signer)
		if err != nil {
			return err
		}
		cg.serviceAccount.accountKey.SequenceNumber++
		cg.step++

		err = cg.flowClient.SendTransaction(context.Background(), *addKeysTx)
		examples.Handle(err)

		accountCreationTxRes := waitForFinalized(context.Background(), cg.flowClient, addKeysTx.ID())
		examples.Handle(accountCreationTxRes.Error)

		fmt.Println("load generator step 0 done")
		return nil
	}
	// setup accounts
	if cg.step == 1 {
		fmt.Println("load generator step 1 started")
		allTxWG := sync.WaitGroup{}
		for i := 0; i < cg.numberOfAccounts; i++ {
			privKey := examples.RandomPrivateKey()
			accountKey := flowsdk.NewAccountKey().
				FromPrivateKey(privKey).
				SetHashAlgo(crypto.SHA3_256).
				SetWeight(flowsdk.AccountKeyWeightThreshold)
			signer := crypto.NewInMemorySigner(privKey, accountKey.HashAlgo)
			createAccountScript, err := templates.CreateAccount([]*flowsdk.AccountKey{accountKey}, nil)
			// Generate an account creation script
			examples.Handle(err)
			createAccountTx := flowsdk.NewTransaction().
				SetReferenceBlockID(ref.ID).
				SetScript(createAccountScript).
				AddAuthorizer(*cg.serviceAccount.address).
				SetProposalKey(*cg.serviceAccount.address, i+1, 0).
				SetPayer(*cg.serviceAccount.address)

			cg.serviceAccount.signerLock.Lock()
			err = createAccountTx.SignEnvelope(*cg.serviceAccount.address, i+1, cg.serviceAccount.signer)
			if err != nil {
				return err
			}
			cg.serviceAccount.signerLock.Unlock()

			err = cg.flowClient.SendTransaction(context.Background(), *createAccountTx)
			examples.Handle(err)
			allTxWG.Add(1)

			cg.txTracker.addTx(createAccountTx.ID(),
				*cg.serviceAccount.address,
				nil,
				func(_ flowsdk.Identifier, res *flowsdk.TransactionResult) {
					for _, event := range res.Events {
						if event.Type == flowsdk.EventAccountCreated {
							accountCreatedEvent := flowsdk.AccountCreatedEvent(event)
							accountAddress := accountCreatedEvent.Address()
							newAcc := newFlowAccount(&accountAddress, accountKey, signer)
							cg.accounts = append(cg.accounts, newAcc)
							fmt.Println("account added")
						}
					}
					allTxWG.Done()
				}, nil, nil, 30)

			// accountCreationTxRes := waitForFinalized(context.Background(), cg.flowClient, createAccountTxID)
			// examples.Handle(accountCreationTxRes.Error)

			// wait group
			// unlock on finalized

			fmt.Println("<<<", i)
			// // Successful Tx, increment sequence number

			fmt.Println(">>", i)
		}
		allTxWG.Wait()
		cg.step++
		fmt.Println("load generator step 2 done")
	}
	// wait for all
	time.Sleep(time.Second * 150)
	// TODO else do the transfers
	return nil
}

// languageEncodeBytes converts a byte slice to a comma-separated list of uint8 integers.
func languageEncodeBytes(b []byte) string {
	if len(b) == 0 {
		return "[]"
	}

	return strings.Join(strings.Fields(fmt.Sprintf("%d", b)), ",")
}

// TODO use context deadlines

func waitForFinalized(ctx context.Context, c *client.Client, id flowsdk.Identifier) *flowsdk.TransactionResult {
	result, err := c.GetTransactionResult(ctx, id)
	// Handle(err)
	fmt.Printf("Waiting for transaction %s to be finalized...\n", id)
	errCount := 0
	for result == nil || (result.Status != flowsdk.TransactionStatusFinalized && result.Status != flowsdk.TransactionStatusSealed) || len(result.Events) == 0 {
		time.Sleep(time.Second)
		result, err = c.GetTransactionResult(ctx, id)
		if err != nil {
			fmt.Print("x")
			errCount++
			if errCount >= 10 {
				return &flowsdk.TransactionResult{
					Error: err,
				}
			}
		} else {
			fmt.Print(".")
		}
		// Handle(err)
	}
	fmt.Println()
	fmt.Printf("Transaction %s finalized\n", id)

	return result
}

func main() {

	serviceAccountAddressHex := "8c5303eaa26202d6"
	fungibleTokenAddressHex := "9a0766d93b6608b7"
	flowTokenAddressHex := "7e60df042a9c0868"

	serviceAccountPrivateKeyBytes, err := hex.DecodeString(unittest.ServiceAccountPrivateKeyHex)
	if err != nil {
		panic("error while hex decoding hardcoded root key")
	}

	// RLP decode the key
	ServiceAccountPrivateKey, err := flow.DecodeAccountPrivateKey(serviceAccountPrivateKeyBytes)
	if err != nil {
		panic("error while decoding hardcoded root key bytes")
	}

	// get the private key string
	priv := hex.EncodeToString(ServiceAccountPrivateKey.PrivateKey.Encode())

	flowClient, err := client.New("localhost:3569", grpc.WithInsecure())
	lg, err := NewLoadGenerator(flowClient, priv, serviceAccountAddressHex, fungibleTokenAddressHex, flowTokenAddressHex, 10)
	if err != nil {
		panic(err)
	}
	lg.Next()
	lg.Next()
}