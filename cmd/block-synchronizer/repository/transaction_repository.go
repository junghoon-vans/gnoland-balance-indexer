package repository

import (
	"gnoland-balance-indexer/pkg/database"
	"gnoland-balance-indexer/pkg/models"
)

type TransactionRepository interface {
	SaveTransaction(tx *models.Transaction) error
}

type transactionRepository struct {
	db *database.Database
}

func NewTransactionRepository(db *database.Database) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) SaveTransaction(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}