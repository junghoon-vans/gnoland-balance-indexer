package repository

import (
	"shared/pkg/database"
	"shared/pkg/models"

	"gorm.io/gorm"
)

type BalanceRepository interface {
	GetBalance(address, tokenPath string) (*models.TokenBalance, error)
	CreateBalance(balance *models.TokenBalance) error
	UpdateBalance(balance *models.TokenBalance) error
	GetBalanceInTx(tx *gorm.DB, address, tokenPath string) (*models.TokenBalance, error)
	CreateBalanceInTx(tx *gorm.DB, balance *models.TokenBalance) error
	UpdateBalanceInTx(tx *gorm.DB, balance *models.TokenBalance) error
}

type balanceRepository struct {
	db *database.Database
}

func NewBalanceRepository(db *database.Database) BalanceRepository {
	return &balanceRepository{db: db}
}

func (r *balanceRepository) GetBalance(address, tokenPath string) (*models.TokenBalance, error) {
	var balance models.TokenBalance
	err := r.db.Where("address = ? AND token_path = ?", address, tokenPath).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *balanceRepository) CreateBalance(balance *models.TokenBalance) error {
	return r.db.Create(balance).Error
}

func (r *balanceRepository) UpdateBalance(balance *models.TokenBalance) error {
	return r.db.Save(balance).Error
}

func (r *balanceRepository) GetBalanceInTx(tx *gorm.DB, address, tokenPath string) (*models.TokenBalance, error) {
	var balance models.TokenBalance
	err := tx.Where("address = ? AND token_path = ?", address, tokenPath).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *balanceRepository) CreateBalanceInTx(tx *gorm.DB, balance *models.TokenBalance) error {
	return tx.Create(balance).Error
}

func (r *balanceRepository) UpdateBalanceInTx(tx *gorm.DB, balance *models.TokenBalance) error {
	return tx.Save(balance).Error
}
