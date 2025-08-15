package database

import (
	"fmt"
	"log"
	"time"

	"gnoland-balance-indexer/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
}

func NewPostgresDB() (*Database, error) {
	dbConfig := Load()
	dsn := dbConfig.DatabaseURL()

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Database{db}, nil
}

func (d *Database) AutoMigrate() error {
	log.Println("Running database migrations...")

	return d.DB.AutoMigrate(
		&models.Block{},
		&models.Transaction{},
		&models.TransactionMsg{},
		&models.TransactionEvent{},
		&models.TransactionEventAttr{},
		&models.TokenBalance{},
		&models.TokenTransfer{},
	)
}

func (d *Database) CreateUniqueIndexes() error {
	log.Println("Creating unique indexes...")

	queries := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_token_balance_address_token 
		 ON token_balances (address, token_path)`,
		`CREATE INDEX IF NOT EXISTS idx_token_transfers_from_address 
		 ON token_transfers (from_address)`,
		`CREATE INDEX IF NOT EXISTS idx_token_transfers_to_address 
		 ON token_transfers (to_address)`,
		`CREATE INDEX IF NOT EXISTS idx_token_transfers_token_path 
		 ON token_transfers (token_path)`,
		`CREATE INDEX IF NOT EXISTS idx_transaction_events_block_height 
		 ON transaction_events (transaction_id)`,
	}

	for _, query := range queries {
		if err := d.DB.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}