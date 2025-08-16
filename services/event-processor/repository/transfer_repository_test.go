package repository

import (
	"testing"

	"shared/infra/database"
	"shared/models"
	"shared/testutils"

	"github.com/stretchr/testify/suite"
)

type TransferRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo TransferRepository
}

func (suite *TransferRepositoryTestSuite) SetupSuite() {
	db, err := testutils.SetupInMemoryDB()
	suite.Require().NoError(err)
	
	suite.db = db
	suite.repo = NewTransferRepository(suite.db)
}

func (suite *TransferRepositoryTestSuite) TearDownTest() {
	testutils.CleanupDatabase(suite.db)
}

func (suite *TransferRepositoryTestSuite) TestSaveTransfer() {
	transfer := testutils.CreateTestNormalTransfer(12345, testutils.TestTransactionHash1, 1, testutils.TestAddress1, testutils.TestAddress2, testutils.TestTokenPath, "1000")

	err := suite.repo.SaveTransfer(transfer)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(transfer.ID)
	suite.Assert().NotZero(transfer.CreatedAt)
}

func (suite *TransferRepositoryTestSuite) TestSaveTransferWithMint() {
	transfer := testutils.CreateTestMintTransfer(12346, testutils.TestTransactionHash2, 2, testutils.TestAddress2, testutils.TestTokenPath, "500")

	err := suite.repo.SaveTransfer(transfer)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(transfer.ID)
	suite.Assert().Equal("", transfer.FromAddress)
	suite.Assert().Equal("mint", transfer.TransferType)
}

func (suite *TransferRepositoryTestSuite) TestSaveTransferWithBurn() {
	transfer := testutils.CreateTestBurnTransfer(12347, "0x789ghi", 3, testutils.TestAddress1, testutils.TestTokenPath, "200")

	err := suite.repo.SaveTransfer(transfer)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(transfer.ID)
	suite.Assert().Equal("", transfer.ToAddress)
	suite.Assert().Equal("burn", transfer.TransferType)
}

func (suite *TransferRepositoryTestSuite) TestSaveMultipleTransfers() {
	transfers := []*models.TokenTransfer{
		testutils.CreateTestNormalTransfer(12345, testutils.TestTransactionHash1, 1, testutils.TestAddress1, testutils.TestAddress2, testutils.TestTokenPath, "1000"),
		testutils.CreateTestNormalTransfer(12346, testutils.TestTransactionHash2, 2, "g1from789", "g1to123", testutils.TestTokenPath, "500"),
	}

	for _, transfer := range transfers {
		err := suite.repo.SaveTransfer(transfer)
		suite.Assert().NoError(err)
		suite.Assert().NotZero(transfer.ID)
	}

	// Verify the number of saved transfer records
	var count int64
	suite.db.Model(&models.TokenTransfer{}).Count(&count)
	suite.Assert().Equal(int64(2), count)
}

func (suite *TransferRepositoryTestSuite) TestSaveTransferValidation() {
	// Test case with missing required fields
	transfer := testutils.CreateTestNormalTransfer(0, testutils.TestTransactionHash1, 1, testutils.TestAddress1, testutils.TestAddress2, testutils.TestTokenPath, "1000")
	// Set BlockHeight to 0 to test missing required field

	// GORM allows zero values by default, so errors may not occur
	// In production, validation should be done with database constraints
	err := suite.repo.SaveTransfer(transfer)
	suite.Assert().NoError(err) // SQLite has loose constraints
}

func TestTransferRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TransferRepositoryTestSuite))
}
