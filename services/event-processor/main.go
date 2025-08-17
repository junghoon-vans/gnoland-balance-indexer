package main

import (
	"context"
	repository2 "event-processor/internal/domain/repository"
	service2 "event-processor/internal/domain/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shared/pkg/database"
	"shared/pkg/queue"
)

func main() {
	log.Println("Starting Event Processor...")

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqsClient, err := queue.NewSQSClient()
	if err != nil {
		log.Fatalf("Failed to create SQS client: %v", err)
	}

	// Initialize repositories
	balanceRepo := repository2.NewBalanceRepository(db)
	transferRepo := repository2.NewTransferRepository(db)
	processedEventRepo := repository2.NewProcessedEventRepository(db)

	// Initialize services
	balanceService := service2.NewBalanceService(db, balanceRepo)
	tokenEventHandler := service2.NewTokenEventHandler(transferRepo, processedEventRepo, balanceService)
	messageProcessor := service2.NewMessageProcessor(sqsClient, tokenEventHandler, 10)
	processorService := service2.NewProcessorService(messageProcessor)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := processorService.Start(ctx); err != nil {
			log.Fatalf("Event processor error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down Event Processor...")

	cancel()
	time.Sleep(5 * time.Second)
	log.Println("Event Processor stopped")
}
