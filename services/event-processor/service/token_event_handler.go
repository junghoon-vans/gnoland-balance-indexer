package service

import (
	"context"
	"fmt"

	"event-processor/repository"
	"event-processor/utils"
	"shared/infra/queue"
	"shared/models"
)

type TokenEventHandler interface {
	ProcessTokenEvent(ctx context.Context, event *queue.TokenEvent) error
}

type tokenEventHandler struct {
	transferRepo   repository.TransferRepository
	balanceService BalanceService
}

func NewTokenEventHandler(
	transferRepo repository.TransferRepository,
	balanceService BalanceService,
) TokenEventHandler {
	return &tokenEventHandler{
		transferRepo:   transferRepo,
		balanceService: balanceService,
	}
}

func (h *tokenEventHandler) ProcessTokenEvent(ctx context.Context, event *queue.TokenEvent) error {
	// Save transfer record
	transfer := &models.TokenTransfer{
		BlockHeight:  event.BlockHeight,
		TxHash:       event.TxHash,
		EventID:      event.EventID,
		FromAddress:  event.FromAddress,
		ToAddress:    event.ToAddress,
		TokenPath:    event.PkgPath,
		Amount:       event.Amount,
		TransferType: utils.GetTransferType(event.Func),
	}

	if err := h.transferRepo.SaveTransfer(transfer); err != nil {
		return fmt.Errorf("failed to save transfer: %w", err)
	}

	// Update balances
	if err := h.balanceService.UpdateBalances(ctx, event); err != nil {
		return fmt.Errorf("failed to update balances: %w", err)
	}

	return nil
}
