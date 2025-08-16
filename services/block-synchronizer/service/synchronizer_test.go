package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"block-synchronizer/dto"
	"shared/models"
	"shared/testutils"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock BlockService for testing
type MockBlockServiceForSynchronizer struct {
	mock.Mock
}

func (m *MockBlockServiceForSynchronizer) GetLatestBlockHeight() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBlockServiceForSynchronizer) SyncBlockRange(ctx context.Context, startHeight, endHeight int64) error {
	args := m.Called(ctx, startHeight, endHeight)
	return args.Error(0)
}

func (m *MockBlockServiceForSynchronizer) ProcessBlock(ctx context.Context, gqlBlock *dto.GraphQLBlock) error {
	args := m.Called(ctx, gqlBlock)
	return args.Error(0)
}

type SynchronizerServiceTestSuite struct {
	suite.Suite
	blockRepo    *MockBlockRepository
	blockService *MockBlockServiceForSynchronizer
	service      SynchronizerService
}

func (suite *SynchronizerServiceTestSuite) SetupTest() {
	suite.blockRepo = new(MockBlockRepository)
	suite.blockService = new(MockBlockServiceForSynchronizer)

	suite.service = NewSynchronizerService(
		suite.blockRepo,
		suite.blockService,
	)
}

func (suite *SynchronizerServiceTestSuite) TearDownTest() {
	testutils.AssertMockExpectations(suite.T(), &suite.blockRepo.Mock, &suite.blockService.Mock)
}

func (suite *SynchronizerServiceTestSuite) TestStart_WithExistingBlocks() {
	ctx, cancel := context.WithTimeout(testutils.CreateTestContext(), 100*time.Millisecond)
	defer cancel()

	// Mock existing block
	existingBlock := testutils.CreateTestBlock(100, "existing-hash")

	// Mock GetLastBlock to return existing block
	suite.blockRepo.On("GetLastBlock").Return(existingBlock, nil).Once()

	// Mock GetLatestBlockHeight to return same height (no new blocks)
	suite.blockService.On("GetLatestBlockHeight").Return(int64(100), nil).Maybe()

	err := suite.service.Start(ctx)
	suite.Assert().NoError(err)
}

func (suite *SynchronizerServiceTestSuite) TestStart_WithNoExistingBlocks() {
	ctx, cancel := context.WithTimeout(testutils.CreateTestContext(), 100*time.Millisecond)
	defer cancel()

	// Mock no existing blocks
	suite.blockRepo.On("GetLastBlock").Return((*models.Block)(nil), errors.New("no blocks found")).Once()

	// Mock GetLatestBlockHeight for backfill
	suite.blockService.On("GetLatestBlockHeight").Return(int64(10), nil).Once()

	// Mock SyncBlockRange for backfill
	suite.blockService.On("SyncBlockRange", mock.Anything, int64(1), int64(10)).Return(nil).Once()

	err := suite.service.Start(ctx)
	suite.Assert().NoError(err)
}

func (suite *SynchronizerServiceTestSuite) TestStart_BackfillError() {
	ctx, cancel := context.WithTimeout(testutils.CreateTestContext(), 100*time.Millisecond)
	defer cancel()

	// Mock no existing blocks
	suite.blockRepo.On("GetLastBlock").Return((*models.Block)(nil), errors.New("no blocks found")).Once()

	// Mock GetLatestBlockHeight for backfill
	suite.blockService.On("GetLatestBlockHeight").Return(int64(10), nil).Once()

	// Mock SyncBlockRange error for backfill
	suite.blockService.On("SyncBlockRange", mock.Anything, int64(1), int64(10)).Return(errors.New("sync error")).Once()

	// Should continue despite backfill error
	err := suite.service.Start(ctx)
	suite.Assert().NoError(err)
}

func (suite *SynchronizerServiceTestSuite) TestStart_GetLatestBlockHeightError() {
	ctx, cancel := context.WithTimeout(testutils.CreateTestContext(), 100*time.Millisecond)
	defer cancel()

	// Mock existing block
	existingBlock := testutils.CreateTestBlock(100, "existing-hash")

	// Mock GetLastBlock to return existing block
	suite.blockRepo.On("GetLastBlock").Return(existingBlock, nil).Once()

	// Mock GetLatestBlockHeight error for backfill
	suite.blockService.On("GetLatestBlockHeight").Return(int64(0), errors.New("height error")).Maybe()

	// Should continue despite height error
	err := suite.service.Start(ctx)
	suite.Assert().NoError(err)
}

func (suite *SynchronizerServiceTestSuite) TestStart_ContextCancellation() {
	ctx, cancel := context.WithCancel(testutils.CreateTestContext())
	cancel() // Cancel immediately

	// Mock existing block
	existingBlock := testutils.CreateTestBlock(100, "existing-hash")

	// Mock GetLastBlock to return existing block
	suite.blockRepo.On("GetLastBlock").Return(existingBlock, nil).Once()

	// Mock GetLatestBlockHeight for backfill
	suite.blockService.On("GetLatestBlockHeight").Return(int64(100), nil).Maybe()

	err := suite.service.Start(ctx)
	suite.Assert().NoError(err)
}

func (suite *SynchronizerServiceTestSuite) TestStart_WithNewBlocks() {
	ctx, cancel := context.WithTimeout(testutils.CreateTestContext(), 200*time.Millisecond)
	defer cancel()

	// Mock existing block
	existingBlock := testutils.CreateTestBlock(100, "existing-hash")

	// Mock GetLastBlock to return existing block
	suite.blockRepo.On("GetLastBlock").Return(existingBlock, nil).Once()

	// Mock GetLatestBlockHeight for backfill (no backfill needed)
	suite.blockService.On("GetLatestBlockHeight").Return(int64(100), nil).Once()

	// Mock GetLatestBlockHeight for sync (new blocks available)
	suite.blockService.On("GetLatestBlockHeight").Return(int64(105), nil).Maybe()

	// Mock SyncBlockRange for new blocks
	suite.blockService.On("SyncBlockRange", ctx, int64(101), int64(105)).Return(nil).Maybe()

	err := suite.service.Start(ctx)
	suite.Assert().NoError(err)
}

func TestSynchronizerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SynchronizerServiceTestSuite))
}
