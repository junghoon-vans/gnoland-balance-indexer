package service

import (
	"context"
	"event-processor/internal/domain/repository"
	"fmt"
	"math/big"

	"shared/pkg/database"
	"shared/pkg/models"
	"shared/pkg/queue"

	"gorm.io/gorm"
)

type BalanceService interface {
	UpdateBalances(ctx context.Context, event *queue.TokenEvent) error
	UpdateBalanceAtomic(ctx context.Context, address, tokenPath, amount string) error
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

func (s *balanceService) UpdateBalances(ctx context.Context, event *queue.TokenEvent) error {
	amount := new(big.Int)
	if _, ok := amount.SetString(event.Amount, 10); !ok {
		return fmt.Errorf("invalid amount format: %s", event.Amount)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		switch event.Type {
		case queue.TransferTypeMint:
			return s.increaseBalance(tx, event.ToAddress, event.PkgPath, amount)
		case queue.TransferTypeBurn:
			return s.decreaseBalance(tx, event.FromAddress, event.PkgPath, amount)
		case queue.TransferTypeTransfer:
			if err := s.decreaseBalance(tx, event.FromAddress, event.PkgPath, amount); err != nil {
				return err
			}
			return s.increaseBalance(tx, event.ToAddress, event.PkgPath, amount)
		default:
			return fmt.Errorf("unknown transfer type: %s", event.Type)
		}
	})
}

func (s *balanceService) increaseBalance(tx *gorm.DB, address, tokenPath string, amount *big.Int) error {
	balance, err := s.balanceRepo.GetBalanceInTx(tx, address, tokenPath)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
	balance.Amount = newAmount.String()

	return s.balanceRepo.UpdateBalanceInTx(tx, balance)
}

func (s *balanceService) decreaseBalance(tx *gorm.DB, address, tokenPath string, amount *big.Int) error {
	balance, err := s.balanceRepo.GetBalanceInTx(tx, address, tokenPath)
	if err != nil {
		return err
	}

	currentAmount := new(big.Int)
	if _, ok := currentAmount.SetString(balance.Amount, 10); !ok {
		return fmt.Errorf("invalid current balance format: %s", balance.Amount)
	}

	if currentAmount.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance: current %s, required %s", currentAmount.String(), amount.String())
	}

	newAmount := new(big.Int).Sub(currentAmount, amount)
	balance.Amount = newAmount.String()

	return s.balanceRepo.UpdateBalanceInTx(tx, balance)
}

// UpdateBalanceAtomic updates balance using the existing transaction-based approach
func (s *balanceService) UpdateBalanceAtomic(ctx context.Context, address, tokenPath, amount string) error {
	amountBig := new(big.Int)
	if _, ok := amountBig.SetString(amount, 10); !ok {
		return fmt.Errorf("invalid amount format: %s", amount)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if amountBig.Sign() >= 0 {
			return s.increaseBalance(tx, address, tokenPath, amountBig)
		} else {
			// Convert to positive for decrease operation
			absAmount := new(big.Int).Abs(amountBig)
			return s.decreaseBalance(tx, address, tokenPath, absAmount)
		}
	})
}
