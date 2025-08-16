package service

import (
	"context"
	"fmt"
	"log"

	"balance-api/dto"
	"balance-api/repository"
	"balance-api/utils"
	"shared/infra/cache"
	"shared/models"
)

type TransferService interface {
	GetTransferHistory(req dto.TransferHistoryRequest) (*dto.TransferHistoryResponse, error)
}

type transferService struct {
	transferRepo repository.TransferRepository
	cache        cache.Cache
}

func NewTransferService(transferRepo repository.TransferRepository, cache cache.Cache) TransferService {
	return &transferService{
		transferRepo: transferRepo,
		cache:        cache,
	}
}

func (s *transferService) GetTransferHistory(req dto.TransferHistoryRequest) (*dto.TransferHistoryResponse, error) {
	ctx := context.Background()
	limit := req.Limit
	if limit <= 0 {
		limit = 1000 // default limit
	}

	// Generate cache key
	var cacheKey string
	if req.Address != "" {
		if !utils.IsValidAddress(req.Address) {
			return nil, fmt.Errorf("invalid address format")
		}
		cacheKey = GenerateTransferKey(req.Address, limit)
	} else {
		cacheKey = GenerateTransferKey("", limit)
	}

	// Try to get from cache first
	var cachedResponse dto.TransferHistoryResponse
	if err := s.cache.Get(ctx, cacheKey, &cachedResponse); err == nil {
		log.Printf("Cache hit for transfer history: %s", cacheKey)
		return &cachedResponse, nil
	} else if err != cache.ErrCacheMiss {
		log.Printf("Cache error for transfer history %s: %v", cacheKey, err)
	}

	// Cache miss, fetch from database
	var transfers []models.TokenTransfer
	var err error

	if req.Address != "" {
		transfers, err = s.transferRepo.GetTransfersByAddress(req.Address, limit)
	} else {
		transfers, err = s.transferRepo.GetAllTransfers(limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch transfer history: %w", err)
	}

	// Convert to DTO
	transferInfos := make([]dto.TransferInfo, 0, len(transfers))
	for _, transfer := range transfers {
		transferInfos = append(transferInfos, dto.TransferInfo{
			FromAddress: transfer.FromAddress,
			ToAddress:   transfer.ToAddress,
			TokenPath:   transfer.TokenPath,
			Amount:      transfer.Amount,
		})
	}

	response := &dto.TransferHistoryResponse{Transfers: transferInfos}

	// Cache the result with 2 minute TTL (transfer history changes less frequently)
	if err := s.cache.Set(ctx, cacheKey, response, TransferHistoryTTL); err != nil {
		log.Printf("Failed to cache transfer history for %s: %v", cacheKey, err)
	}

	return response, nil
}
