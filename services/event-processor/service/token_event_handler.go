package service

import (
	"context"
	"fmt"
	"log"

	"event-processor/repository"
	"event-processor/utils"
	"shared/infra/queue"
	"shared/models"
)

type TokenEventHandler interface {
	ProcessTokenEvent(ctx context.Context, event *queue.TokenEvent) error
}

type tokenEventHandler struct {
	transferRepo       repository.TransferRepository
	processedEventRepo repository.ProcessedEventRepository
	balanceService     BalanceService
}

func NewTokenEventHandler(
	transferRepo repository.TransferRepository,
	processedEventRepo repository.ProcessedEventRepository,
	balanceService BalanceService,
) TokenEventHandler {
	return &tokenEventHandler{
		transferRepo:       transferRepo,
		processedEventRepo: processedEventRepo,
		balanceService:     balanceService,
	}
}

func (h *tokenEventHandler) ProcessTokenEvent(ctx context.Context, event *queue.TokenEvent) error {
	// Generate unique event identifier
	eventIdentifier := fmt.Sprintf("%s-%d", event.TxHash, event.EventID)

	// Check if event is already processed (idempotency check)
	isProcessed, err := h.processedEventRepo.IsEventProcessed(eventIdentifier)
	if err != nil {
		return fmt.Errorf("failed to check if event is processed: %w", err)
	}

	if isProcessed {
		log.Printf("Event %s already processed, skipping (idempotent)", eventIdentifier)
		return nil
	}

	// Create transfer record
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

	// Save transfer record (with unique constraint protection)
	if err := h.transferRepo.SaveTransfer(transfer); err != nil {
		return fmt.Errorf("failed to save transfer: %w", err)
	}

	// Update balances
	if err := h.balanceService.UpdateBalances(ctx, event); err != nil {
		return fmt.Errorf("failed to update balances: %w", err)
	}

	// Mark event as processed for idempotency
	processedEvent := &models.ProcessedEvent{
		EventIdentifier: eventIdentifier,
		TxHash:          event.TxHash,
		EventID:         event.EventID,
		BlockHeight:     event.BlockHeight,
	}

	if err := h.processedEventRepo.MarkEventProcessed(processedEvent); err != nil {
		log.Printf("Warning: Failed to mark event %s as processed: %v", eventIdentifier, err)
		// Don't fail the whole operation if we can't mark as processed
		// The unique constraint on transfers table will still prevent duplicates
	}

	log.Printf("Successfully processed event %s (tx: %s, event: %d)", eventIdentifier, event.TxHash, event.EventID)
	return nil
}
