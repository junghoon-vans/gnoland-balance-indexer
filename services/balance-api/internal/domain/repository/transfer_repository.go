package repository

import (
	"shared/pkg/database"
	"shared/pkg/models"
)

type TransferRepository interface {
	GetTransfersByAddress(address string, limit int) ([]models.TokenTransfer, error)
	GetAllTransfers(limit int) ([]models.TokenTransfer, error)
}

type transferRepository struct {
	db *database.Database
}

func NewTransferRepository(db *database.Database) TransferRepository {
	return &transferRepository{db: db}
}

func (r *transferRepository) GetTransfersByAddress(address string, limit int) ([]models.TokenTransfer, error) {
	var transfers []models.TokenTransfer
	query := r.db.Model(&models.TokenTransfer{}).
		Where("from_address = ? OR to_address = ?", address, address).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&transfers).Error
	return transfers, err
}

func (r *transferRepository) GetAllTransfers(limit int) ([]models.TokenTransfer, error) {
	var transfers []models.TokenTransfer
	query := r.db.Model(&models.TokenTransfer{}).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&transfers).Error
	return transfers, err
}
