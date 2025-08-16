package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"block-synchronizer/dto"
	"block-synchronizer/repository"
	"shared/infra/queue"
	"shared/models"
)

type EventService interface {
	ProcessEvent(ctx context.Context, txID uint, gqlTx *dto.GraphQLTransaction, gqlEvent *dto.GraphQLEvent) error
}

type eventService struct {
	eventRepo repository.EventRepository
	sqsClient *queue.SQSClient
}

func NewEventService(
	eventRepo repository.EventRepository,
	sqsClient *queue.SQSClient,
) EventService {
	return &eventService{
		eventRepo: eventRepo,
		sqsClient: sqsClient,
	}
}

func (s *eventService) ProcessEvent(ctx context.Context, txID uint, gqlTx *dto.GraphQLTransaction, gqlEvent *dto.GraphQLEvent) error {
	if gqlEvent == nil {
		return fmt.Errorf("event cannot be nil")
	}

	event := &models.TransactionEvent{
		TransactionID: txID,
		Type:          gqlEvent.Type,
		Func:          gqlEvent.Func,
		PkgPath:       gqlEvent.PkgPath,
	}

	if err := s.eventRepo.SaveEvent(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	attrs := make(map[string]string)
	for _, attr := range gqlEvent.Attrs {
		eventAttr := &models.TransactionEventAttr{
			EventID: event.ID,
			Key:     attr.Key,
			Value:   attr.Value,
		}

		if err := s.eventRepo.SaveEventAttr(eventAttr); err != nil {
			return fmt.Errorf("failed to save event attribute: %w", err)
		}

		attrs[attr.Key] = attr.Value
	}

	if s.isTokenTransferEvent(gqlEvent.Type, gqlEvent.Func, attrs) {
		return s.sendTokenEventToQueue(ctx, gqlTx, event, attrs)
	}

	return nil
}

func (s *eventService) isTokenTransferEvent(eventType, funcName string, attrs map[string]string) bool {
	// Check if it's a token transfer event
	if eventType == "Transfer" || strings.Contains(funcName, "Transfer") {
		// Verify required attributes exist
		_, hasFrom := attrs["from"]
		_, hasTo := attrs["to"]
		_, hasValue := attrs["value"]
		return hasFrom && hasTo && hasValue
	}
	return false
}

func (s *eventService) isValidAddress(addr string) bool {
	if addr == "" {
		return false
	}
	// Basic address validation - adjust pattern as needed for your blockchain
	matched, _ := regexp.MatchString(`^g[a-zA-Z0-9]{38,}$`, addr)
	return matched
}

func (s *eventService) sendTokenEventToQueue(ctx context.Context, gqlTx *dto.GraphQLTransaction, event *models.TransactionEvent, attrs map[string]string) error {
	fromAddr := attrs["from"]
	toAddr := attrs["to"]
	amount := attrs["value"]

	if _, err := strconv.ParseInt(amount, 10, 64); err != nil {
		return fmt.Errorf("invalid amount format: %s", amount)
	}

	if !s.isValidAddress(fromAddr) && fromAddr != "" {
		return fmt.Errorf("invalid from address: %s", fromAddr)
	}

	if !s.isValidAddress(toAddr) && toAddr != "" {
		return fmt.Errorf("invalid to address: %s", toAddr)
	}

	tokenEvent := &queue.TokenEvent{
		ID:          fmt.Sprintf("%s-%d", gqlTx.Hash, event.ID),
		BlockHeight: gqlTx.BlockHeight,
		TxHash:      gqlTx.Hash,
		EventID:     event.ID,
		Type:        event.Type,
		Func:        event.Func,
		PkgPath:     event.PkgPath,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      amount,
		Timestamp:   time.Now(),
		Attributes:  attrs,
	}

	return s.sqsClient.SendMessage(tokenEvent)
}
