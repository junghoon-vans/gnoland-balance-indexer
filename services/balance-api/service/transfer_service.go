package service

import (
	"fmt"

	"balance-api/dto"
	"balance-api/repository"
	"balance-api/utils"
	"shared/models"
)

type TransferService interface {
	GetTransferHistory(req dto.TransferHistoryRequest) (*dto.TransferHistoryResponse, error)
}

type transferService struct {
	transferRepo repository.TransferRepository
}

func NewTransferService(transferRepo repository.TransferRepository) TransferService {
	return &transferService{
		transferRepo: transferRepo,
	}
}

func (s *transferService) GetTransferHistory(req dto.TransferHistoryRequest) (*dto.TransferHistoryResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 1000 // default limit
	}

	var transfers []models.TokenTransfer
	var err error

	if req.Address != "" {
		if !utils.IsValidAddress(req.Address) {
			return nil, fmt.Errorf("invalid address format")
		}
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

	return &dto.TransferHistoryResponse{Transfers: transferInfos}, nil
}
