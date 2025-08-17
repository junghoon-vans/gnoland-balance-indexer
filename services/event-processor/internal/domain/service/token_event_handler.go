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
	ProcessTokenEvent(ctx context.Context, event *queue.TokenEvent) error
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

	// Update balances only (transfer record is now saved by block-synchronizer)
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
	}

	log.Printf("Successfully processed event %s (tx: %s, event: %d)", eventIdentifier, event.TxHash, event.EventID)
	return nil
}

// ProcessBalanceUpdate processes a single balance update message (new FIFO approach)
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

	// Update balance directly using atomic operation
	if err := h.balanceService.UpdateBalanceAtomic(ctx, update.Address, update.TokenPath, update.Amount); err != nil {
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
		// Don't fail the operation - the atomic balance update already succeeded
	}

	log.Printf("Successfully processed balance update %s (address: %s, amount: %s)",
		eventIdentifier, update.Address, update.Amount)
	return nil
}
