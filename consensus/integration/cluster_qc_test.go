package integration_test

import (
	"context"
	"testing"

	"github.com/onflow/flow-core-contracts/lib/go/contracts"
	"github.com/onflow/flow-core-contracts/lib/go/templates"
	"github.com/onflow/flow-emulator/server/backend"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/test"

	sdk "github.com/onflow/flow-go-sdk"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
)

func TestQuroumCertificate(t *testing.T) {
	backend := backend.New(nil, backend.Emulator{})

	// amulator environment
	env := templates.Environment{
		FlowTokenAddress: "",
	}

	// create account keys to deploy
	accountKeys := test.AccountKeyGenerator()

	// Create new keys for the QC account and deploy
	QCAccountKey, QCSigner := accountKeys.NewWithSigner()
	QCCode := contracts.FlowQC()

	// get latest sealed block
	latestBlock, err := backend.GetLatestBlock(context.Background(), true)
	if err != nil {
		return sdk.Address{}, err
	}

	// deploy contract to emulator
	deployContractTx := sdktemplates.CreateAccount([]*flow.AccountKey{QCAccountKey}, []sdktemplates.Contract{
		{
			Name:   "FlowEpochClusterQC",
			Source: string(QCCode),
		},
	})
	deployContractTx.SetGasLimit(10000).
		SetReferenceBlockID(sdk.Identifier(latestBlock.ID()))
}
