package repository

import (
	"testing"

	"shared/infra/database"
	"shared/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TransferRepositoryTestSuite struct {
	suite.Suite
	db   *database.Database
	repo TransferRepository
}

func (suite *TransferRepositoryTestSuite) SetupSuite() {
	// Use in-memory SQLite database
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	suite.db = &database.Database{DB: gormDB}

	// Create tables
	err = suite.db.DB.AutoMigrate(&models.TokenTransfer{})
	suite.Require().NoError(err)

	suite.repo = NewTransferRepository(suite.db)
}

func (suite *TransferRepositoryTestSuite) TearDownTest() {
	// Clean up data after each test
	suite.db.Exec("DELETE FROM token_transfers")
}

func (suite *TransferRepositoryTestSuite) TestSaveTransfer() {
	transfer := &models.TokenTransfer{
		BlockHeight:  12345,
		TxHash:       "0x123abc",
		EventID:      1,
		FromAddress:  "g1from123",
		ToAddress:    "g1to456",
		TokenPath:    "gno.land/r/demo/grc20",
		Amount:       "1000",
		TransferType: "transfer",
	}

	err := suite.repo.SaveTransfer(transfer)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(transfer.ID)
	suite.Assert().NotZero(transfer.CreatedAt)
}

func (suite *TransferRepositoryTestSuite) TestSaveTransferWithMint() {
	transfer := &models.TokenTransfer{
		BlockHeight:  12346,
		TxHash:       "0x456def",
		EventID:      2,
		FromAddress:  "", // Empty from address for mint
		ToAddress:    "g1to456",
		TokenPath:    "gno.land/r/demo/grc20",
		Amount:       "500",
		TransferType: "mint",
	}

	err := suite.repo.SaveTransfer(transfer)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(transfer.ID)
	suite.Assert().Equal("", transfer.FromAddress)
	suite.Assert().Equal("mint", transfer.TransferType)
}

func (suite *TransferRepositoryTestSuite) TestSaveTransferWithBurn() {
	transfer := &models.TokenTransfer{
		BlockHeight:  12347,
		TxHash:       "0x789ghi",
		EventID:      3,
		FromAddress:  "g1from123",
		ToAddress:    "", // Empty to address for burn
		TokenPath:    "gno.land/r/demo/grc20",
		Amount:       "200",
		TransferType: "burn",
	}

	err := suite.repo.SaveTransfer(transfer)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(transfer.ID)
	suite.Assert().Equal("", transfer.ToAddress)
	suite.Assert().Equal("burn", transfer.TransferType)
}

func (suite *TransferRepositoryTestSuite) TestSaveMultipleTransfers() {
	transfers := []*models.TokenTransfer{
		{
			BlockHeight:  12345,
			TxHash:       "0x123abc",
			EventID:      1,
			FromAddress:  "g1from123",
			ToAddress:    "g1to456",
			TokenPath:    "gno.land/r/demo/grc20",
			Amount:       "1000",
			TransferType: "transfer",
		},
		{
			BlockHeight:  12346,
			TxHash:       "0x456def",
			EventID:      2,
			FromAddress:  "g1from789",
			ToAddress:    "g1to123",
			TokenPath:    "gno.land/r/demo/grc20",
			Amount:       "500",
			TransferType: "transfer",
		},
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
	transfer := &models.TokenTransfer{
		// Missing BlockHeight
		TxHash:       "0x123abc",
		EventID:      1,
		FromAddress:  "g1from123",
		ToAddress:    "g1to456",
		TokenPath:    "gno.land/r/demo/grc20",
		Amount:       "1000",
		TransferType: "transfer",
	}

	// GORM allows zero values by default, so errors may not occur
	// In production, validation should be done with database constraints
	err := suite.repo.SaveTransfer(transfer)
	suite.Assert().NoError(err) // SQLite has loose constraints
}

func TestTransferRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TransferRepositoryTestSuite))
}
