package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"balance-api/dto"
	"balance-api/repository"
	"balance-api/utils"
	"shared/infra/cache"
	"shared/models"
)

type BalanceService interface {
	GetTokenBalances(req dto.BalanceRequest) (*dto.BalanceResponse, error)
	GetTokenBalancesByPath(tokenPath string, address string) (*dto.AccountBalancesResponse, error)
}

type balanceService struct {
	balanceRepo repository.BalanceRepository
	cache       cache.Cache
}

func NewBalanceService(balanceRepo repository.BalanceRepository, cache cache.Cache) BalanceService {
	return &balanceService{
		balanceRepo: balanceRepo,
		cache:       cache,
	}
}

func (s *balanceService) GetTokenBalances(req dto.BalanceRequest) (*dto.BalanceResponse, error) {
	ctx := context.Background()

	if req.Address != "" {
		if !utils.IsValidAddress(req.Address) {
			return nil, fmt.Errorf("invalid address format")
		}

		// Try to get from cache first
		cacheKey := GenerateBalanceAddressKey(req.Address)
		var cachedResponse dto.BalanceResponse
		if err := s.cache.Get(ctx, cacheKey, &cachedResponse); err == nil {
			log.Printf("Cache hit for address: %s", req.Address)
			return &cachedResponse, nil
		} else if err != cache.ErrCacheMiss {
			log.Printf("Cache error for address %s: %v", req.Address, err)
		}

		// Cache miss, fetch from database
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

		response := &dto.BalanceResponse{Balances: tokenBalances}

		// Cache the result
		if err := s.cache.Set(ctx, cacheKey, response, BalanceAddressTTL); err != nil {
			log.Printf("Failed to cache address balances for %s: %v", req.Address, err)
		}

		return response, nil
	}

	// Get all balances (with shorter cache TTL due to frequent updates)
	cacheKey := GenerateBalanceAddressKey("")
	var cachedResponse dto.BalanceResponse
	if err := s.cache.Get(ctx, cacheKey, &cachedResponse); err == nil {
		log.Printf("Cache hit for all balances")
		return &cachedResponse, nil
	} else if err != cache.ErrCacheMiss {
		log.Printf("Cache error for all balances: %v", err)
	}

	// Cache miss, fetch from database
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

	response := &dto.BalanceResponse{Balances: tokenBalances}

	// Cache with shorter TTL for aggregate data
	if err := s.cache.Set(ctx, cacheKey, response, BalanceAllTTL); err != nil {
		log.Printf("Failed to cache all balances: %v", err)
	}

	return response, nil
}

func (s *balanceService) GetTokenBalancesByPath(tokenPath string, address string) (*dto.AccountBalancesResponse, error) {
	ctx := context.Background()

	// Decode URL-encoded token path
	decodedTokenPath := strings.ReplaceAll(tokenPath, "%2F", "/")

	// Generate cache key
	if address != "" && !utils.IsValidAddress(address) {
		return nil, fmt.Errorf("invalid address format")
	}
	cacheKey := GenerateBalanceTokenKey(decodedTokenPath, address)

	// Try to get from cache first
	var cachedResponse dto.AccountBalancesResponse
	if err := s.cache.Get(ctx, cacheKey, &cachedResponse); err == nil {
		log.Printf("Cache hit for token path: %s, address: %s", decodedTokenPath, address)
		return &cachedResponse, nil
	} else if err != cache.ErrCacheMiss {
		log.Printf("Cache error for token path %s, address %s: %v", decodedTokenPath, address, err)
	}

	// Cache miss, fetch from database
	var balances []models.TokenBalance
	var err error

	if address != "" {
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

	response := &dto.AccountBalancesResponse{AccountBalances: accountBalances}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, response, BalanceTokenTTL); err != nil {
		log.Printf("Failed to cache token path balances for %s: %v", decodedTokenPath, err)
	}

	return response, nil
}
