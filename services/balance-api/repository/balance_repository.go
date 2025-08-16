package repository

import (
	"shared/infra/database"
	"shared/models"
)

type BalanceRepository interface {
	GetBalancesByAddress(address string) ([]models.TokenBalance, error)
	GetAllBalances() ([]models.TokenBalance, error)
	GetBalancesByTokenPath(tokenPath string) ([]models.TokenBalance, error)
	GetBalancesByTokenPathAndAddress(tokenPath, address string) ([]models.TokenBalance, error)
}

type balanceRepository struct {
	db *database.Database
}

func NewBalanceRepository(db *database.Database) BalanceRepository {
	return &balanceRepository{db: db}
}

func (r *balanceRepository) GetBalancesByAddress(address string) ([]models.TokenBalance, error) {
	var balances []models.TokenBalance
	err := r.db.Model(&models.TokenBalance{}).
		Where("address = ? AND amount != '0'", address).
		Find(&balances).Error
	return balances, err
}

func (r *balanceRepository) GetAllBalances() ([]models.TokenBalance, error) {
	var balances []models.TokenBalance
	err := r.db.Model(&models.TokenBalance{}).
		Where("amount != '0'").
		Find(&balances).Error
	return balances, err
}

func (r *balanceRepository) GetBalancesByTokenPath(tokenPath string) ([]models.TokenBalance, error) {
	var balances []models.TokenBalance
	err := r.db.Model(&models.TokenBalance{}).
		Where("token_path = ? AND amount != '0'", tokenPath).
		Find(&balances).Error
	return balances, err
}

func (r *balanceRepository) GetBalancesByTokenPathAndAddress(tokenPath, address string) ([]models.TokenBalance, error) {
	var balances []models.TokenBalance
	err := r.db.Model(&models.TokenBalance{}).
		Where("token_path = ? AND address = ? AND amount != '0'", tokenPath, address).
		Find(&balances).Error
	return balances, err
}
