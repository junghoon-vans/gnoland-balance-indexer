package service

import (
	"balance-api/internal/api/dto"
	"balance-api/internal/domain/repository"
	"balance-api/pkg/utils"
	"fmt"
	"strconv"
	"strings"

	"shared/pkg/models"
)

type BalanceService interface {
	GetTokenBalances(req dto.BalanceRequest) (*dto.BalanceResponse, error)
	GetTokenBalancesByPath(tokenPath string, address string) (*dto.AccountBalancesResponse, error)
}

type balanceService struct {
	balanceRepo repository.BalanceRepository
}

func NewBalanceService(balanceRepo repository.BalanceRepository) BalanceService {
	return &balanceService{
		balanceRepo: balanceRepo,
	}
}

func (s *balanceService) GetTokenBalances(req dto.BalanceRequest) (*dto.BalanceResponse, error) {
	if req.Address != "" {
		if !utils.IsValidAddress(req.Address) {
			return nil, fmt.Errorf("invalid address format")
		}

		balances, err := s.balanceRepo.GetBalancesByAddress(req.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch balances: %w", err)
		}

		tokenBalances := make([]dto.TokenBalanceInfo, 0, len(balances))
		for _, balance := range balances {
			tokenBalances = append(tokenBalances, dto.TokenBalanceInfo{
				TokenPath: balance.TokenPath,
				Amount:    balance.Amount,
			})
		}

		return &dto.BalanceResponse{Balances: tokenBalances}, nil
	}

	// Get all balances and aggregate by token path
	balances, err := s.balanceRepo.GetAllBalances()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balances: %w", err)
	}

	tokenMap := make(map[string]string)
	for _, balance := range balances {
		if currentAmount, exists := tokenMap[balance.TokenPath]; exists {
			currentInt, _ := strconv.ParseInt(currentAmount, 10, 64)
			balanceInt, _ := strconv.ParseInt(balance.Amount, 10, 64)
			tokenMap[balance.TokenPath] = strconv.FormatInt(currentInt+balanceInt, 10)
		} else {
			tokenMap[balance.TokenPath] = balance.Amount
		}
	}

	tokenBalances := make([]dto.TokenBalanceInfo, 0)
	for tokenPath, amount := range tokenMap {
		if amount != "0" {
			tokenBalances = append(tokenBalances, dto.TokenBalanceInfo{
				TokenPath: tokenPath,
				Amount:    amount,
			})
		}
	}

	return &dto.BalanceResponse{Balances: tokenBalances}, nil
}

func (s *balanceService) GetTokenBalancesByPath(tokenPath string, address string) (*dto.AccountBalancesResponse, error) {
	// Decode URL-encoded token path
	decodedTokenPath := strings.ReplaceAll(tokenPath, "%2F", "/")

	var balances []models.TokenBalance
	var err error

	if address != "" {
		if !utils.IsValidAddress(address) {
			return nil, fmt.Errorf("invalid address format")
		}
		balances, err = s.balanceRepo.GetBalancesByTokenPathAndAddress(decodedTokenPath, address)
	} else {
		balances, err = s.balanceRepo.GetBalancesByTokenPath(decodedTokenPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch balances: %w", err)
	}

	accountBalances := make([]dto.AccountBalanceInfo, 0, len(balances))
	for _, balance := range balances {
		accountBalances = append(accountBalances, dto.AccountBalanceInfo{
			Address:   balance.Address,
			TokenPath: balance.TokenPath,
			Amount:    balance.Amount,
		})
	}

	return &dto.AccountBalancesResponse{AccountBalances: accountBalances}, nil
}
