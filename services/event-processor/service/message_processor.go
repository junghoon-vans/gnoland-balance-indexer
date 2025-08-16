package service

import (
	"context"
	"fmt"
	"log"

	"shared/infra/queue"
)

type MessageProcessor interface {
	ProcessMessages(ctx context.Context) error
}

type messageProcessor struct {
	sqsClient        *queue.SQSClient
	tokenEventHandler TokenEventHandler
	maxMessages      int64
}

func NewMessageProcessor(
	sqsClient *queue.SQSClient,
	tokenEventHandler TokenEventHandler,
	maxMessages int64,
) MessageProcessor {
	return &messageProcessor{
		sqsClient:        sqsClient,
		tokenEventHandler: tokenEventHandler,
		maxMessages:      maxMessages,
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
		if err := p.tokenEventHandler.ProcessTokenEvent(ctx, &msg.Body); err != nil {
			log.Printf("Error processing token event %s: %v", msg.Body.ID, err)
			continue
		}
		log.Printf("Successfully processed token event %s", msg.Body.ID)
	}

	return nil
}