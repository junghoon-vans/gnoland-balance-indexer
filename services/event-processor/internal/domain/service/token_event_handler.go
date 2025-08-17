package service

import (
	"context"
	"event-processor/internal/domain/repository"
	"fmt"
	"log"

	"shared/pkg/models"
	"shared/pkg/queue"
)

type TokenEventHandler interface {
	ProcessBalanceUpdate(ctx context.Context, update *queue.BalanceUpdateMessage) error
}

type tokenEventHandler struct {
	processedEventRepo repository.ProcessedEventRepository
	balanceService     BalanceService
}

func NewTokenEventHandler(
	processedEventRepo repository.ProcessedEventRepository,
	balanceService BalanceService,
) TokenEventHandler {
	return &tokenEventHandler{
		processedEventRepo: processedEventRepo,
		balanceService:     balanceService,
	}
}

// ProcessBalanceUpdate processes a single balance update message
func (h *tokenEventHandler) ProcessBalanceUpdate(ctx context.Context, update *queue.BalanceUpdateMessage) error {
	// Generate unique event identifier for idempotency
	eventIdentifier := fmt.Sprintf("%s-%s-%d", update.TxHash, update.Address, update.EventID)

	// Check if this specific balance update is already processed
	isProcessed, err := h.processedEventRepo.IsEventProcessed(eventIdentifier)
	if err != nil {
		return fmt.Errorf("failed to check if balance update is processed: %w", err)
	}

	if isProcessed {
		log.Printf("Balance update %s already processed, skipping (idempotent)", eventIdentifier)
		return nil
	}

	if err := h.balanceService.UpdateBalance(ctx, update.Address, update.TokenPath, update.Amount); err != nil {
		return fmt.Errorf("failed to update balance atomically: %w", err)
	}

	// Mark this balance update as processed for idempotency
	processedEvent := &models.ProcessedEvent{
		EventIdentifier: eventIdentifier,
		TxHash:          update.TxHash,
		EventID:         update.EventID,
		BlockHeight:     update.BlockHeight,
	}

	if err := h.processedEventRepo.MarkEventProcessed(processedEvent); err != nil {
		log.Printf("Warning: Failed to mark balance update %s as processed: %v", eventIdentifier, err)
	}

	log.Printf("Successfully processed balance update %s (address: %s, amount: %s)",
		eventIdentifier, update.Address, update.Amount)
	return nil
}
