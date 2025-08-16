package service

import (
	"context"
	"log"
	"time"
)

type ProcessorService interface {
	Start(ctx context.Context) error
}

type processorService struct {
	messageProcessor MessageProcessor
	processInterval  time.Duration
}

func NewProcessorService(
	messageProcessor MessageProcessor,
) ProcessorService {
	return &processorService{
		messageProcessor: messageProcessor,
		processInterval:  5 * time.Second,
	}
}

func (s *processorService) Start(ctx context.Context) error {
	log.Println("Event Processor started")

	ticker := time.NewTicker(s.processInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Event Processor stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.messageProcessor.ProcessMessages(ctx); err != nil {
				log.Printf("Error processing messages: %v", err)
			}
		}
	}
}