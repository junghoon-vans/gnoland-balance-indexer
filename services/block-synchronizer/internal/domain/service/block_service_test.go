package service

import (
	"block-synchronizer/internal/api/dto"
	"context"
	"errors"
	"testing"
	"time"

	"shared/pkg/graphql"
	"shared/pkg/models"
	"shared/pkg/testutils"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock repositories using testify/mock
type MockBlockRepository struct {
	mock.Mock
}

func (m *MockBlockRepository) GetLastBlock() (*models.Block, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Block), args.Error(1)
}

func (m *MockBlockRepository) SaveBlock(block *models.Block) error {
	args := m.Called(block)
	return args.Error(0)
}

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) SaveTransaction(tx *models.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) SaveEvent(event *models.TransactionEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventRepository) SaveEventAttr(attr *models.TransactionEventAttr) error {
	args := m.Called(attr)
	return args.Error(0)
}

type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) ProcessEvent(event *models.TransactionEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) ProcessTransactions(ctx context.Context, blockHeight int64) error {
	args := m.Called(ctx, blockHeight)
	return args.Error(0)
}

func (m *MockTransactionService) ProcessTransaction(ctx context.Context, gqlTx *dto.GraphQLTransaction) error {
	args := m.Called(ctx, gqlTx)
	return args.Error(0)
}

type MockGraphQLClient struct {
	mock.Mock
}

func (m *MockGraphQLClient) Query(query string, variables map[string]interface{}) (*graphql.GraphQLResponse, error) {
	args := m.Called(query, variables)
	return args.Get(0).(*graphql.GraphQLResponse), args.Error(1)
}

// Test suite
type BlockServiceTestSuite struct {
	suite.Suite
	blockRepo          *MockBlockRepository
	transactionRepo    *MockTransactionRepository
	eventRepo          *MockEventRepository
	gqlClient          *MockGraphQLClient
	transactionService *MockTransactionService
	service            BlockService
}

func (suite *BlockServiceTestSuite) SetupTest() {
	suite.blockRepo = new(MockBlockRepository)
	suite.transactionRepo = new(MockTransactionRepository)
	suite.eventRepo = new(MockEventRepository)
	suite.gqlClient = new(MockGraphQLClient)
	suite.transactionService = new(MockTransactionService)

	suite.service = NewBlockService(
		suite.blockRepo,
		suite.transactionRepo,
		suite.eventRepo,
		suite.gqlClient,
		suite.transactionService,
	)
}

func (suite *BlockServiceTestSuite) TearDownTest() {
	testutils.AssertMockExpectations(suite.T(), &suite.blockRepo.Mock, &suite.transactionRepo.Mock, &suite.eventRepo.Mock, &suite.gqlClient.Mock)
}

func (suite *BlockServiceTestSuite) TestGetLatestBlockHeight() {
	// Test case: GraphQL returns height
	response := &graphql.GraphQLResponse{
		Data: []byte(`{"latestBlockHeight": 12345}`),
	}

	suite.gqlClient.On("Query", mock.AnythingOfType("string"), mock.Anything).Return(response, nil).Once()

	height, err := suite.service.GetLatestBlockHeight()
	suite.Assert().NoError(err)
	suite.Assert().Equal(int64(12345), height)
}

func (suite *BlockServiceTestSuite) TestGetLatestBlockHeightWhenNoBlock() {
	// Test case: GraphQL returns error
	suite.gqlClient.On("Query", mock.AnythingOfType("string"), mock.Anything).Return(&graphql.GraphQLResponse{}, errors.New("graphql error")).Once()

	height, err := suite.service.GetLatestBlockHeight()
	suite.Assert().Error(err)
	suite.Assert().Equal(int64(0), height)
}

func (suite *BlockServiceTestSuite) TestProcessBlock() {
	gqlBlock := &dto.GraphQLBlock{
		Hash:     "test-hash",
		Height:   1,
		Time:     time.Now().Format(time.RFC3339),
		NumTxs:   5,
		TotalTxs: 10,
	}

	// Set up expectations
	suite.blockRepo.On("SaveBlock", mock.MatchedBy(func(b *models.Block) bool {
		return b.Height == 1 && b.Hash == "test-hash"
	})).Return(nil).Once()
	suite.transactionService.On("ProcessTransactions", mock.Anything, int64(1)).Return(nil).Once()

	ctx := testutils.CreateTestContext()
	err := suite.service.ProcessBlock(ctx, gqlBlock)
	suite.Assert().NoError(err)
}

func (suite *BlockServiceTestSuite) TestProcessBlockWithInvalidTime() {
	gqlBlock := &dto.GraphQLBlock{
		Hash:     "test-hash",
		Height:   1,
		Time:     "invalid-time-format",
		NumTxs:   5,
		TotalTxs: 10,
	}

	ctx := testutils.CreateTestContext()
	err := suite.service.ProcessBlock(ctx, gqlBlock)
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "parsing time")
}

func (suite *BlockServiceTestSuite) TestProcessBlockWithSaveError() {
	gqlBlock := &dto.GraphQLBlock{
		Hash:     "test-hash",
		Height:   1,
		Time:     time.Now().Format(time.RFC3339),
		NumTxs:   5,
		TotalTxs: 10,
	}

	// Set up expectation for SaveBlock to return error
	suite.blockRepo.On("SaveBlock", mock.MatchedBy(func(b *models.Block) bool {
		return b.Height == 1
	})).Return(errors.New("database error")).Once()

	ctx := testutils.CreateTestContext()
	err := suite.service.ProcessBlock(ctx, gqlBlock)
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "database error")
}

func TestBlockServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BlockServiceTestSuite))
}
