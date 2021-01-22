package epochs

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/client/convert"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"
	access "github.com/onflow/flow/protobuf/go/flow/access"

	hotstuff "github.com/onflow/flow-go/consensus/hotstuff/mocks"
	hotstuffmodel "github.com/onflow/flow-go/consensus/hotstuff/model"
	"github.com/onflow/flow-go/model/cluster"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/epochs"
	modulemock "github.com/onflow/flow-go/module/mock"
	"github.com/onflow/flow-go/utils/unittest"

	clusterstate "github.com/onflow/flow-go/state/cluster"
	protomock "github.com/onflow/flow-go/state/protocol/mock"
)

type ClusterNode struct {
	NodeID  flow.Identifier
	Key     *sdk.AccountKey
	Address sdk.Address
	Voter   module.ClusterRootQCVoter
}

// TestQuroumCertificate tests one Epoch of the EpochClusterQC contract
func (s *ClusterEpochTestSuite) TestQuorumCertificate() {

	// initial cluster and total node count
	clusterCount := 3
	nodeCount := 30

	s.SetupTest()

	// create clustering with x clusters with x*y nodes
	clustering, nodes := s.CreateClusterList(clusterCount, nodeCount)

	// create initial ClusterNodes list
	clusterNodes := make([]*ClusterNode, clusterCount)

	// mock the epoch object to return counter 0 and clustering as our clusterList
	epoch := &protomock.Epoch{}
	epoch.On("Counter").Return(1, nil)
	epoch.On("Clustering").Return(clustering, nil)

	for _, node := range nodes {

		// find cluster and create root block
		cluster, _, _ := clustering.ByNodeID(node.NodeID)
		rootBlock := clusterstate.CanonicalRootBlock(1, cluster)

		clusterNode := s.CreateClusterNode(rootBlock, node)
		clusterNodes = append(clusterNodes, clusterNode)
	}

	err := s.PublishVoter()
	require.NoError(s.T(), err)

	err = s.StartVoting(clustering, clusterCount, nodeCount)
	require.NoError(s.T(), err)

	err = s.CreateVoterResource(clustering)
	require.NoError(s.T(), err)

	// cast vote to qc contract for each node
	for _, node := range clusterNodes {
		err = node.Voter.Vote(context.Background(), epoch)
		require.NoError(s.T(), err)
	}

	err = s.StopVoting()
	require.NoError(s.T(), err)

	// TODO: check contract and see if required results are there
}

// CreateClusterNode ...
func (s *ClusterEpochTestSuite) CreateClusterNode(rootBlock *cluster.Block, me *flow.Identity) *ClusterNode {

	key, signer := test.AccountKeyGenerator().NewWithSigner()

	// create account on emualted chain
	address, err := s.blockchain.CreateAccount([]*sdk.AccountKey{key}, []sdktemplates.Contract{})
	require.NoError(s.T(), err)

	// create a mock of QCContractClient to submit vote and check if voted on the emulated chain
	rpcClient := &MockRPCClient{}
	rpcClient.On("GetAccount", mock.Anything, mock.Anything).
		Return(func(_ context.Context, address sdk.Address) access.GetAccountResponse {
			account, err := s.blockchain.GetAccount(address)
			require.NoError(s.T(), err)
			return access.GetAccountResponse{Account: convert.AccountToMessage(*account)}
		}, func(_ context.Context, address sdk.Address) error {
			return nil
		})
	rpcClient.On("SendTransaction", mock.Anything, mock.Anything).
		Return(func(_ context.Context, tx *sdk.Transaction) error {
			return s.Submit(tx)
		})
	rpcClient.On("GetLatestBlock", mock.Anything, mock.Anything).
		Return(func(_ context.Context, _ bool) access.BlockResponse {
			block, err := s.blockchain.GetLatestBlock()
			require.NoError(s.T(), err)
			blockID := block.ID()

			var id sdk.Identifier
			copy(id[:], blockID[:])

			sdkblock := sdk.Block{
				BlockHeader: sdk.BlockHeader{ID: id},
			}

			msg, err := convert.BlockToMessage(sdkblock)
			require.NoError(s.T(), err)

			return access.BlockResponse{Block: msg}
		})
	rpcClient.On("GetTransactionResult", mock.Anything, mock.Anything).
		Return(func(_ context.Context, txID flow.Identifier) access.TransactionResultResponse {

			var id sdk.Identifier
			copy(id[:], txID[:])

			result, err := s.blockchain.GetTransactionResult(id)
			require.NoError(s.T(), err)

			msg, err := convert.TransactionResultToMessage(*result)
			require.NoError(s.T(), err)

			return *msg
		})

	flowClient := client.NewFromRPCClient(rpcClient)

	client, err := epochs.NewQCContractClient(flowClient, me.NodeID, address.String(), uint(key.Index), s.qcAddress.String(), signer)
	require.NoError(s.T(), err)

	local := &modulemock.Local{}
	local.On("NodeID").Return(me.NodeID)

	vote := hotstuffmodel.VoteFromFlow(me.NodeID, unittest.IdentifierFixture(), 0, unittest.SignatureFixture())

	hotSigner := &hotstuff.Signer{}
	hotSigner.On("CreateVote", mock.Anything).Return(vote, nil)

	snapshot := &protomock.Snapshot{}
	snapshot.On("Phase").Return(flow.EpochPhaseSetup, nil)

	state := &protomock.State{}
	state.On("CanonicalRootBlock").Return(rootBlock)
	state.On("Final").Return(snapshot)

	// create QC voter object to be used for voting for the root QC contract
	voter := epochs.NewRootQCVoter(zerolog.Logger{}, local, hotSigner, state, client)

	// create node and set node
	node := &ClusterNode{
		NodeID:  me.NodeID,
		Key:     key,
		Address: address,
		Voter:   voter,
	}

	return node
}
