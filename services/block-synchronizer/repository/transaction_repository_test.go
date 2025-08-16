package repository

import (
	"testing"

	"shared/infra/database"
	"shared/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TransactionRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo TransactionRepository
}

func (suite *TransactionRepositoryTestSuite) SetupSuite() {
	// Use in-memory SQLite database
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	suite.db = &database.Database{DB: gormDB}

	// Create tables
	err = suite.db.DB.AutoMigrate(&models.Transaction{}, &models.TransactionMsg{}, &models.TransactionEvent{})
	suite.Require().NoError(err)

	suite.repo = NewTransactionRepository(suite.db)
}

func (suite *TransactionRepositoryTestSuite) TearDownTest() {
	// Clean up data after each test
	suite.db.Exec("DELETE FROM transaction_msgs")
	suite.db.Exec("DELETE FROM transaction_events")
	suite.db.Exec("DELETE FROM transactions")
}

func (suite *TransactionRepositoryTestSuite) TestSaveTransaction() {
	tx := &models.Transaction{
		Hash:        "0x123abc",
		Index:       0,
		BlockHeight: 12345,
		Success:     true,
		GasWanted:   100000,
		GasUsed:     50000,
		Memo:        "test transaction",
	}

	err := suite.repo.SaveTransaction(tx)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(tx.ID)
	suite.Assert().NotZero(tx.CreatedAt)

	// Verify transaction was saved
	var savedTx models.Transaction
	err = suite.db.First(&savedTx, tx.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal("0x123abc", savedTx.Hash)
	suite.Assert().Equal(0, savedTx.Index)
	suite.Assert().Equal(int64(12345), savedTx.BlockHeight)
	suite.Assert().Equal(true, savedTx.Success)
	suite.Assert().Equal(int64(100000), savedTx.GasWanted)
	suite.Assert().Equal(int64(50000), savedTx.GasUsed)
	suite.Assert().Equal("test transaction", savedTx.Memo)
}

func (suite *TransactionRepositoryTestSuite) TestSaveTransactionWithoutMemo() {
	tx := &models.Transaction{
		Hash:        "0x456def",
		Index:       1,
		BlockHeight: 12346,
		Success:     false,
		GasWanted:   200000,
		GasUsed:     150000,
		// No memo
	}

	err := suite.repo.SaveTransaction(tx)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(tx.ID)

	// Verify transaction was saved
	var savedTx models.Transaction
	err = suite.db.First(&savedTx, tx.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal("0x456def", savedTx.Hash)
	suite.Assert().Equal(false, savedTx.Success)
	suite.Assert().Equal("", savedTx.Memo)
}

func (suite *TransactionRepositoryTestSuite) TestSaveTransactionWithDuplicateHash() {
	tx1 := &models.Transaction{
		Hash:        "0x123abc",
		Index:       0,
		BlockHeight: 12345,
		Success:     true,
		GasWanted:   100000,
		GasUsed:     50000,
	}

	tx2 := &models.Transaction{
		Hash:        "0x123abc", // Same hash
		Index:       1,
		BlockHeight: 12346,
		Success:     true,
		GasWanted:   200000,
		GasUsed:     100000,
	}

	// Save first transaction
	err := suite.repo.SaveTransaction(tx1)
	suite.Assert().NoError(err)

	// Try to save second transaction with same hash - should fail
	err = suite.repo.SaveTransaction(tx2)
	suite.Assert().Error(err) // Should fail due to unique constraint
}

func (suite *TransactionRepositoryTestSuite) TestSaveMultipleTransactions() {
	transactions := []*models.Transaction{
		{
			Hash:        "0x123abc",
			Index:       0,
			BlockHeight: 12345,
			Success:     true,
			GasWanted:   100000,
			GasUsed:     50000,
			Memo:        "first transaction",
		},
		{
			Hash:        "0x456def",
			Index:       1,
			BlockHeight: 12345,
			Success:     false,
			GasWanted:   200000,
			GasUsed:     150000,
			Memo:        "second transaction",
		},
		{
			Hash:        "0x789ghi",
			Index:       0,
			BlockHeight: 12346,
			Success:     true,
			GasWanted:   300000,
			GasUsed:     200000,
			Memo:        "third transaction",
		},
	}

	// Save all transactions
	for _, tx := range transactions {
		err := suite.repo.SaveTransaction(tx)
		suite.Assert().NoError(err)
		suite.Assert().NotZero(tx.ID)
	}

	// Verify all transactions were saved
	var count int64
	suite.db.Model(&models.Transaction{}).Count(&count)
	suite.Assert().Equal(int64(3), count)
}

func (suite *TransactionRepositoryTestSuite) TestSaveTransactionWithZeroValues() {
	tx := &models.Transaction{
		Hash:        "0x000000",
		Index:       0,
		BlockHeight: 0, // Zero block height
		Success:     false,
		GasWanted:   0, // Zero gas wanted
		GasUsed:     0, // Zero gas used
		Memo:        "",
	}

	// GORM allows zero values by default
	err := suite.repo.SaveTransaction(tx)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(tx.ID)

	// Verify transaction was saved with zero values
	var savedTx models.Transaction
	err = suite.db.First(&savedTx, tx.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal(int64(0), savedTx.BlockHeight)
	suite.Assert().Equal(int64(0), savedTx.GasWanted)
	suite.Assert().Equal(int64(0), savedTx.GasUsed)
}

func (suite *TransactionRepositoryTestSuite) TestSaveTransactionWithLargeValues() {
	tx := &models.Transaction{
		Hash:        "0xabcdef123456789",
		Index:       999,
		BlockHeight: 9999999999,
		Success:     true,
		GasWanted:   9999999999,
		GasUsed:     8888888888,
		Memo:        "This is a very long memo field that contains a lot of text to test how the database handles longer strings and whether there are any issues with storage or retrieval of such data.",
	}

	err := suite.repo.SaveTransaction(tx)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(tx.ID)

	// Verify transaction was saved with large values
	var savedTx models.Transaction
	err = suite.db.First(&savedTx, tx.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal(999, savedTx.Index)
	suite.Assert().Equal(int64(9999999999), savedTx.BlockHeight)
	suite.Assert().Equal(int64(9999999999), savedTx.GasWanted)
	suite.Assert().Equal(int64(8888888888), savedTx.GasUsed)
	suite.Assert().Contains(savedTx.Memo, "very long memo field")
}

func TestTransactionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionRepositoryTestSuite))
}
