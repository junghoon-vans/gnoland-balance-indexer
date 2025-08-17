package testutils

import (
	"shared/infra/database"
	"shared/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupInMemoryDB creates an in-memory SQLite database for testing
func SetupInMemoryDB() (*database.Database, error) {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db := &database.Database{DB: gormDB}

	// Auto-migrate all models
	err = db.DB.AutoMigrate(
		&models.Block{},
		&models.Transaction{},
		&models.TransactionMsg{},
		&models.TransactionEvent{},
		&models.TransactionEventAttr{},
		&models.TokenBalance{},
		&models.TokenTransfer{},
		&models.ProcessedEvent{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// SetupInMemoryDBOrPanic is an alias for SetupInMemoryDB for test compatibility
func SetupInMemoryDBOrPanic() *database.Database {
	db, err := SetupInMemoryDB()
	if err != nil {
		panic(err)
	}
	return db
}

// CleanupDatabase cleans all tables in the test database
func CleanupDatabase(db *database.Database) {
	db.Exec("DELETE FROM processed_events")
	db.Exec("DELETE FROM token_transfers")
	db.Exec("DELETE FROM token_balances")
	db.Exec("DELETE FROM transaction_event_attrs")
	db.Exec("DELETE FROM transaction_events")
	db.Exec("DELETE FROM transaction_msgs")
	db.Exec("DELETE FROM transactions")
	db.Exec("DELETE FROM blocks")
}
