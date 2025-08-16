package repository

import (
	"shared/infra/database"
	"shared/models"
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
	return r.db.Create(transfer).Error
}
