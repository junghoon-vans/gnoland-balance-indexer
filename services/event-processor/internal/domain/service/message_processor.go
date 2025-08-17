package service

import (
	"context"
	"fmt"
	"log"

	"shared/pkg/queue"
)

type MessageProcessor interface {
	ProcessMessages(ctx context.Context) error
}

type messageProcessor struct {
	sqsClient         *queue.SQSClient
	tokenEventHandler TokenEventHandler
	maxMessages       int64
}

func NewMessageProcessor(
	sqsClient *queue.SQSClient,
	tokenEventHandler TokenEventHandler,
	maxMessages int64,
) MessageProcessor {
	return &messageProcessor{
		sqsClient:         sqsClient,
		tokenEventHandler: tokenEventHandler,
		maxMessages:       maxMessages,
	}
}

func (p *messageProcessor) ProcessMessages(ctx context.Context) error {
	messages, err := p.sqsClient.ReceiveMessages(p.maxMessages)
	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	for _, msg := range messages {
		err := p.tokenEventHandler.ProcessBalanceUpdate(ctx, &msg.Body)
		if err != nil {
			log.Printf("Error processing balance update %s: %v", msg.Body.Address, err)
			continue
		}
		log.Printf("Successfully processed balance update for address %s", msg.Body.Address)

		if err := p.sqsClient.DeleteMessage(msg.ReceiptHandle); err != nil {
			log.Printf("Failed to delete balance_update message %s from queue: %v", msg.MessageID, err)
		} else {
			log.Printf("Successfully deleted balance_update message %s from queue", msg.MessageID)
		}
	}

	return nil
}
