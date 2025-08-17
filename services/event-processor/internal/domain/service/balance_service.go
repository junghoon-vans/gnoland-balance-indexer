package service

import (
	"context"
	"event-processor/internal/domain/repository"
	"fmt"
	"math/big"

	"shared/pkg/database"
	"shared/pkg/models"

	"gorm.io/gorm"
)

type BalanceService interface {
	UpdateBalance(ctx context.Context, address, tokenPath, amount string) error
}

type balanceService struct {
	db          *database.Database
	balanceRepo repository.BalanceRepository
}

func NewBalanceService(db *database.Database, balanceRepo repository.BalanceRepository) BalanceService {
	return &balanceService{
		db:          db,
		balanceRepo: balanceRepo,
	}
}

// UpdateBalance updates balance in a transaction (positive amount increases, negative decreases)
func (s *balanceService) UpdateBalance(ctx context.Context, address, tokenPath, amount string) error {
	amountBig := new(big.Int)
	if _, ok := amountBig.SetString(amount, 10); !ok {
		return fmt.Errorf("invalid amount format: %s", amount)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.updateBalanceInTx(tx, address, tokenPath, amountBig)
	})
}

func (s *balanceService) updateBalanceInTx(tx *gorm.DB, address, tokenPath string, amount *big.Int) error {
	balance, err := s.balanceRepo.GetBalanceInTx(tx, address, tokenPath)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if amount.Sign() < 0 {
				return fmt.Errorf("insufficient balance: no existing balance for negative amount %s", amount.String())
			}
			newBalance := &models.TokenBalance{
				Address:   address,
				TokenPath: tokenPath,
				Amount:    amount.String(),
			}
			return s.balanceRepo.CreateBalanceInTx(tx, newBalance)
		}
		return err
	}

	currentAmount := new(big.Int)
	if _, ok := currentAmount.SetString(balance.Amount, 10); !ok {
		return fmt.Errorf("invalid current balance format: %s", balance.Amount)
	}

	newAmount := new(big.Int).Add(currentAmount, amount)

	if newAmount.Sign() < 0 {
		return fmt.Errorf("insufficient balance: current %s, change %s, result would be %s", currentAmount.String(), amount.String(), newAmount.String())
	}

	balance.Amount = newAmount.String()
	return s.balanceRepo.UpdateBalanceInTx(tx, balance)
}
