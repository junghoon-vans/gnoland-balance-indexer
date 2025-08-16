package repository

import (
	"testing"

	"shared/infra/database"
	"shared/models"
	"shared/testutils"

	"github.com/stretchr/testify/suite"
)

type TransactionRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo TransactionRepository
}

func (suite *TransactionRepositoryTestSuite) SetupSuite() {
	db, err := testutils.SetupInMemoryDB()
	suite.Require().NoError(err)
	
	suite.db = db
	suite.repo = NewTransactionRepository(suite.db)
}

func (suite *TransactionRepositoryTestSuite) TearDownTest() {
	testutils.CleanupDatabase(suite.db)
}

func (suite *TransactionRepositoryTestSuite) TestSaveTransaction() {
	tx := testutils.CreateTestTransaction(testutils.TestTransactionHash1, 12345)

	err := suite.repo.SaveTransaction(tx)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(tx.ID)
	suite.Assert().NotZero(tx.CreatedAt)

	// Verify transaction was saved
	var savedTx models.Transaction
	err = suite.db.First(&savedTx, tx.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal(testutils.TestTransactionHash1, savedTx.Hash)
	suite.Assert().Equal(0, savedTx.Index)
	suite.Assert().Equal(int64(12345), savedTx.BlockHeight)
	suite.Assert().Equal(true, savedTx.Success)
	suite.Assert().Equal(int64(100000), savedTx.GasWanted)
	suite.Assert().Equal(int64(50000), savedTx.GasUsed)
	suite.Assert().Equal("test transaction", savedTx.Memo)
}

func (suite *TransactionRepositoryTestSuite) TestSaveTransactionWithoutMemo() {
	tx := testutils.CreateTestTransaction(testutils.TestTransactionHash2, 12346)
	tx.Index = 1
	tx.Success = false
	tx.GasWanted = 200000
	tx.GasUsed = 150000
	tx.Memo = "" // No memo

	err := suite.repo.SaveTransaction(tx)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(tx.ID)

	// Verify transaction was saved
	var savedTx models.Transaction
	err = suite.db.First(&savedTx, tx.ID).Error
	suite.Assert().NoError(err)
	suite.Assert().Equal(testutils.TestTransactionHash2, savedTx.Hash)
	suite.Assert().Equal(false, savedTx.Success)
	suite.Assert().Equal("", savedTx.Memo)
}

func (suite *TransactionRepositoryTestSuite) TestSaveTransactionWithDuplicateHash() {
	tx1 := testutils.CreateTestTransaction(testutils.TestTransactionHash1, 12345)
	tx2 := testutils.CreateTestTransaction(testutils.TestTransactionHash1, 12346) // Same hash
	tx2.Index = 1
	tx2.GasWanted = 200000
	tx2.GasUsed = 100000

	// Save first transaction
	err := suite.repo.SaveTransaction(tx1)
	suite.Assert().NoError(err)

	// Try to save second transaction with same hash - should fail
	err = suite.repo.SaveTransaction(tx2)
	suite.Assert().Error(err) // Should fail due to unique constraint
}

func (suite *TransactionRepositoryTestSuite) TestSaveMultipleTransactions() {
	tx1 := testutils.CreateTestTransaction(testutils.TestTransactionHash1, 12345)
	tx1.Memo = "first transaction"
	
	tx2 := testutils.CreateTestTransaction(testutils.TestTransactionHash2, 12345)
	tx2.Index = 1
	tx2.Success = false
	tx2.GasWanted = 200000
	tx2.GasUsed = 150000
	tx2.Memo = "second transaction"
	
	tx3 := testutils.CreateTestTransaction("0x789unique", 12346)
	tx3.GasWanted = 300000
	tx3.GasUsed = 200000
	tx3.Memo = "third transaction"
	
	transactions := []*models.Transaction{tx1, tx2, tx3}

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
	tx := testutils.CreateTestTransaction("0x000000", 0) // Zero block height
	tx.Success = false
	tx.GasWanted = 0 // Zero gas wanted
	tx.GasUsed = 0   // Zero gas used
	tx.Memo = ""

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
	tx := testutils.CreateTestTransaction("0xabcdef123456789", 9999999999)
	tx.Index = 999
	tx.GasWanted = 9999999999
	tx.GasUsed = 8888888888
	tx.Memo = "This is a very long memo field that contains a lot of text to test how the database handles longer strings and whether there are any issues with storage or retrieval of such data."

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
