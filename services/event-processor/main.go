package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-processor/repository"
	"event-processor/service"
	"shared/infra/database"
	"shared/infra/queue"
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
	balanceRepo := repository.NewBalanceRepository(db)
	transferRepo := repository.NewTransferRepository(db)

	// Initialize services
	balanceService := service.NewBalanceService(db, balanceRepo)
	tokenEventHandler := service.NewTokenEventHandler(transferRepo, balanceService)
	messageProcessor := service.NewMessageProcessor(sqsClient, tokenEventHandler, 10)
	processorService := service.NewProcessorService(messageProcessor)

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
