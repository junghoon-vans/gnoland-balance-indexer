package service

import (
	dto2 "balance-api/internal/api/dto"
	"balance-api/internal/domain/repository"
	"balance-api/pkg/utils"
	"fmt"

	"shared/pkg/models"
)

type TransferService interface {
	GetTransferHistory(req dto2.TransferHistoryRequest) (*dto2.TransferHistoryResponse, error)
}

type transferService struct {
	transferRepo repository.TransferRepository
}

func NewTransferService(transferRepo repository.TransferRepository) TransferService {
	return &transferService{
		transferRepo: transferRepo,
	}
}

func (s *transferService) GetTransferHistory(req dto2.TransferHistoryRequest) (*dto2.TransferHistoryResponse, error) {
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
	transferInfos := make([]dto2.TransferInfo, 0, len(transfers))
	for _, transfer := range transfers {
		transferInfos = append(transferInfos, dto2.TransferInfo{
			FromAddress: transfer.FromAddress,
			ToAddress:   transfer.ToAddress,
			TokenPath:   transfer.TokenPath,
			Amount:      transfer.Amount,
		})
	}

	return &dto2.TransferHistoryResponse{Transfers: transferInfos}, nil
}
