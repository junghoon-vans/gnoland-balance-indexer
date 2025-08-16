package repository

import (
	"testing"
	"time"

	"gnoland-balance-indexer/pkg/database"
	"gnoland-balance-indexer/pkg/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type BlockRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo BlockRepository
}

func (suite *BlockRepositoryTestSuite) SetupSuite() {
	// Use in-memory SQLite database
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	suite.db = &database.Database{DB: gormDB}

	// Create tables
	err = suite.db.DB.AutoMigrate(&models.Block{})
	suite.Require().NoError(err)

	suite.repo = NewBlockRepository(suite.db)
}

func (suite *BlockRepositoryTestSuite) TearDownTest() {
	// Clean up data after each test
	suite.db.Exec("DELETE FROM blocks")
}

func (suite *BlockRepositoryTestSuite) TestSaveBlock() {
	block := &models.Block{
		Hash:     "0x123abc",
		Height:   12345,
		Time:     time.Now(),
		NumTxs:   5,
		TotalTxs: 100,
	}

	err := suite.repo.SaveBlock(block)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(block.ID)
	suite.Assert().NotZero(block.CreatedAt)
}

func (suite *BlockRepositoryTestSuite) TestGetLastBlock() {
	// Create test blocks with different heights
	blocks := []*models.Block{
		{
			Hash:     "0x123abc",
			Height:   12345,
			Time:     time.Now().Add(-2 * time.Hour),
			NumTxs:   5,
			TotalTxs: 100,
		},
		{
			Hash:     "0x456def",
			Height:   12346,
			Time:     time.Now().Add(-1 * time.Hour),
			NumTxs:   3,
			TotalTxs: 103,
		},
		{
			Hash:     "0x789ghi",
			Height:   12347,
			Time:     time.Now(),
			NumTxs:   7,
			TotalTxs: 110,
		},
	}

	// Save all blocks
	for _, block := range blocks {
		err := suite.repo.SaveBlock(block)
		suite.Require().NoError(err)
	}

	// Get the last block (highest height)
	lastBlock, err := suite.repo.GetLastBlock()
	suite.Assert().NoError(err)
	suite.Assert().NotNil(lastBlock)
	suite.Assert().Equal(int64(12347), lastBlock.Height)
	suite.Assert().Equal("0x789ghi", lastBlock.Hash)
}

func (suite *BlockRepositoryTestSuite) TestGetLastBlockWhenEmpty() {
	// Test when no blocks exist
	lastBlock, err := suite.repo.GetLastBlock()
	suite.Assert().Error(err)
	suite.Assert().Nil(lastBlock)
}

func (suite *BlockRepositoryTestSuite) TestSaveBlockWithDuplicateHash() {
	block1 := &models.Block{
		Hash:     "0x123abc",
		Height:   12345,
		Time:     time.Now(),
		NumTxs:   5,
		TotalTxs: 100,
	}

	block2 := &models.Block{
		Hash:     "0x123abc", // Same hash
		Height:   12346,
		Time:     time.Now(),
		NumTxs:   3,
		TotalTxs: 103,
	}

	// Save first block
	err := suite.repo.SaveBlock(block1)
	suite.Assert().NoError(err)

	// Try to save second block with same hash - should fail
	err = suite.repo.SaveBlock(block2)
	suite.Assert().Error(err) // Should fail due to unique constraint
}

func (suite *BlockRepositoryTestSuite) TestSaveBlockWithDuplicateHeight() {
	block1 := &models.Block{
		Hash:     "0x123abc",
		Height:   12345,
		Time:     time.Now(),
		NumTxs:   5,
		TotalTxs: 100,
	}

	block2 := &models.Block{
		Hash:     "0x456def",
		Height:   12345, // Same height
		Time:     time.Now(),
		NumTxs:   3,
		TotalTxs: 103,
	}

	// Save first block
	err := suite.repo.SaveBlock(block1)
	suite.Assert().NoError(err)

	// Try to save second block with same height - should fail
	err = suite.repo.SaveBlock(block2)
	suite.Assert().Error(err) // Should fail due to unique constraint
}

func TestBlockRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(BlockRepositoryTestSuite))
}