package repository

import (
	"errors"
	"log"

	"shared/pkg/database"
	"shared/pkg/models"

	"gorm.io/gorm"
)

type TransferRepository interface {
	SaveTransfer(transfer *models.TokenTransfer) error
}

type transferRepository struct {
	db *database.Database
}

func NewTransferRepository(db *database.Database) TransferRepository {
	return &transferRepository{db: db}
}

func (r *transferRepository) SaveTransfer(transfer *models.TokenTransfer) error {
	err := r.db.Create(transfer).Error
	if err != nil {
		// Check if it's a duplicate key error (unique constraint violation)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.Printf("Transfer already exists (tx: %s, event: %d), treating as idempotent success",
				transfer.TxHash, transfer.EventID)
			return nil // Treat duplicate as success for idempotency
		}
		return err
	}
	return nil
}
