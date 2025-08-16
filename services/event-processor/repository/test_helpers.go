package repository

import (
	"shared/infra/database"
	"shared/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDatabase creates an in-memory database for testing
func SetupTestDatabase() (*database.Database, error) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db := &database.Database{DB: gormDB}

	// Create tables
	err = db.DB.AutoMigrate(&models.TokenBalance{}, &models.TokenTransfer{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CleanupTestDatabase cleans up all data in the test database
func CleanupTestDatabase(db *database.Database) {
	db.Exec("DELETE FROM token_balances")
	db.Exec("DELETE FROM token_transfers")
}

// CreateTestBalance creates a TokenBalance for testing
func CreateTestBalance(address, tokenPath, amount string) *models.TokenBalance {
	return &models.TokenBalance{
		Address:   address,
		TokenPath: tokenPath,
		Amount:    amount,
	}
}

// CreateTestTransfer creates a TokenTransfer for testing
func CreateTestTransfer(blockHeight int64, txHash string, eventID uint, fromAddr, toAddr, tokenPath, amount, transferType string) *models.TokenTransfer {
	return &models.TokenTransfer{
		BlockHeight:  blockHeight,
		TxHash:       txHash,
		EventID:      eventID,
		FromAddress:  fromAddr,
		ToAddress:    toAddr,
		TokenPath:    tokenPath,
		Amount:       amount,
		TransferType: transferType,
	}
}

// CreateTestMintTransfer creates a mint transfer for testing
func CreateTestMintTransfer(blockHeight int64, txHash string, eventID uint, toAddr, tokenPath, amount string) *models.TokenTransfer {
	return CreateTestTransfer(blockHeight, txHash, eventID, "", toAddr, tokenPath, amount, "mint")
}

// CreateTestBurnTransfer creates a burn transfer for testing
func CreateTestBurnTransfer(blockHeight int64, txHash string, eventID uint, fromAddr, tokenPath, amount string) *models.TokenTransfer {
	return CreateTestTransfer(blockHeight, txHash, eventID, fromAddr, "", tokenPath, amount, "burn")
}

// CreateTestNormalTransfer creates a normal transfer for testing
func CreateTestNormalTransfer(blockHeight int64, txHash string, eventID uint, fromAddr, toAddr, tokenPath, amount string) *models.TokenTransfer {
	return CreateTestTransfer(blockHeight, txHash, eventID, fromAddr, toAddr, tokenPath, amount, "transfer")
}
