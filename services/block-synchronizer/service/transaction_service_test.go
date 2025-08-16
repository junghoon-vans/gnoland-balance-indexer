package service

import (
	"context"
	"errors"
	"testing"

	"block-synchronizer/dto"
	"shared/infra/graphql"
	"shared/models"
	"shared/testutils"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock EventService for testing
type MockEventServiceForTransaction struct {
	mock.Mock
}

func (m *MockEventServiceForTransaction) ProcessEvent(ctx context.Context, txID uint, gqlTx *dto.GraphQLTransaction, gqlEvent *dto.GraphQLEvent) error {
	args := m.Called(ctx, txID, gqlTx, gqlEvent)
	return args.Error(0)
}

type TransactionServiceTestSuite struct {
	suite.Suite
	transactionRepo *MockTransactionRepository
	gqlClient       *MockGraphQLClient
	eventService    *MockEventServiceForTransaction
	service         TransactionService
}

func (suite *TransactionServiceTestSuite) SetupTest() {
	suite.transactionRepo = new(MockTransactionRepository)
	suite.gqlClient = new(MockGraphQLClient)
	suite.eventService = new(MockEventServiceForTransaction)

	suite.service = NewTransactionService(
		suite.transactionRepo,
		suite.gqlClient,
		suite.eventService,
	)
}

func (suite *TransactionServiceTestSuite) TearDownTest() {
	testutils.AssertMockExpectations(suite.T(), &suite.transactionRepo.Mock, &suite.gqlClient.Mock, &suite.eventService.Mock)
}

func (suite *TransactionServiceTestSuite) TestProcessTransactions_Success() {
	ctx := testutils.CreateTestContext()
	blockHeight := int64(100)

	// Mock GraphQL response
	response := &graphql.GraphQLResponse{
		Data: []byte(`{
			"getTransactions": [
				{
					"index": 0,
					"hash": "test-hash-1",
					"success": true,
					"block_height": 100,
					"gas_wanted": 1000,
					"gas_used": 800,
					"memo": "test memo",
					"response": {
						"events": [
							{
								"type": "transfer",
								"func": "Transfer",
								"pkg_path": "gno.land/p/demo/grc20",
								"attrs": [
									{"key": "from", "value": "addr1"},
									{"key": "to", "value": "addr2"}
								]
							}
						]
					}
				}
			]
		}`),
	}

	suite.gqlClient.On("Query", mock.AnythingOfType("string"), mock.MatchedBy(func(vars map[string]interface{}) bool {
		return vars["blockHeight"] == blockHeight
	})).Return(response, nil).Once()

	// Mock transaction save
	suite.transactionRepo.On("SaveTransaction", mock.MatchedBy(func(tx *models.Transaction) bool {
		return tx.Hash == "test-hash-1" && tx.BlockHeight == 100
	})).Return(nil).Once()

	// Mock event processing
	suite.eventService.On("ProcessEvent", ctx, mock.AnythingOfType("uint"), mock.AnythingOfType("*dto.GraphQLTransaction"), mock.AnythingOfType("*dto.GraphQLEvent")).Return(nil).Once()

	err := suite.service.ProcessTransactions(ctx, blockHeight)
	suite.Assert().NoError(err)
}

func (suite *TransactionServiceTestSuite) TestProcessTransactions_GraphQLError() {
	ctx := testutils.CreateTestContext()
	blockHeight := int64(100)

	suite.gqlClient.On("Query", mock.AnythingOfType("string"), mock.Anything).Return(&graphql.GraphQLResponse{}, errors.New("graphql error")).Once()

	err := suite.service.ProcessTransactions(ctx, blockHeight)
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "graphql error")
}

func (suite *TransactionServiceTestSuite) TestProcessTransaction_Success() {
	ctx := testutils.CreateTestContext()
	gqlTx := &dto.GraphQLTransaction{
		Hash:        "test-hash",
		Index:       0,
		BlockHeight: 100,
		Success:     true,
		GasWanted:   1000,
		GasUsed:     800,
		Memo:        "test memo",
		Response: dto.GraphQLTransactionResp{
			Events: []dto.GraphQLEvent{
				{
					Type:    "transfer",
					Func:    "Transfer",
					PkgPath: "gno.land/p/demo/grc20",
					Attrs: []dto.GraphQLEventAttr{
						{Key: "from", Value: "addr1"},
						{Key: "to", Value: "addr2"},
					},
				},
			},
		},
	}

	// Mock transaction save
	suite.transactionRepo.On("SaveTransaction", mock.MatchedBy(func(tx *models.Transaction) bool {
		return tx.Hash == "test-hash" && tx.BlockHeight == 100
	})).Return(nil).Once()

	// Mock event processing
	suite.eventService.On("ProcessEvent", ctx, mock.AnythingOfType("uint"), gqlTx, mock.AnythingOfType("*dto.GraphQLEvent")).Return(nil).Once()

	err := suite.service.ProcessTransaction(ctx, gqlTx)
	suite.Assert().NoError(err)
}

func (suite *TransactionServiceTestSuite) TestProcessTransaction_SaveError() {
	ctx := testutils.CreateTestContext()
	gqlTx := &dto.GraphQLTransaction{
		Hash:        "test-hash",
		Index:       0,
		BlockHeight: 100,
		Success:     true,
		GasWanted:   1000,
		GasUsed:     800,
		Memo:        "test memo",
		Response: dto.GraphQLTransactionResp{
			Events: []dto.GraphQLEvent{},
		},
	}

	// Mock transaction save error
	suite.transactionRepo.On("SaveTransaction", mock.AnythingOfType("*models.Transaction")).Return(errors.New("save error")).Once()

	err := suite.service.ProcessTransaction(ctx, gqlTx)
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "failed to save transaction")
}

func (suite *TransactionServiceTestSuite) TestProcessTransaction_EventProcessingError() {
	ctx := testutils.CreateTestContext()
	gqlTx := &dto.GraphQLTransaction{
		Hash:        "test-hash",
		Index:       0,
		BlockHeight: 100,
		Success:     true,
		GasWanted:   1000,
		GasUsed:     800,
		Memo:        "test memo",
		Response: dto.GraphQLTransactionResp{
			Events: []dto.GraphQLEvent{
				{
					Type:    "transfer",
					Func:    "Transfer",
					PkgPath: "gno.land/p/demo/grc20",
					Attrs: []dto.GraphQLEventAttr{
						{Key: "from", Value: "addr1"},
					},
				},
			},
		},
	}

	// Mock transaction save
	suite.transactionRepo.On("SaveTransaction", mock.AnythingOfType("*models.Transaction")).Return(nil).Once()

	// Mock event processing error (should not fail the transaction)
	suite.eventService.On("ProcessEvent", ctx, mock.AnythingOfType("uint"), gqlTx, mock.AnythingOfType("*dto.GraphQLEvent")).Return(errors.New("event error")).Once()

	err := suite.service.ProcessTransaction(ctx, gqlTx)
	suite.Assert().NoError(err) // Should not fail even if event processing fails
}

func TestTransactionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionServiceTestSuite))
}
