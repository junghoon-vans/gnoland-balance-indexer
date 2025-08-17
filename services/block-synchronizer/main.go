package main

import (
	"block-synchronizer/internal/domain/repository"
	"block-synchronizer/internal/domain/service"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shared/pkg/database"
	"shared/pkg/graphql"
	"shared/pkg/queue"
)

func main() {
	log.Println("Starting Block Synchronizer...")

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqsClient, err := queue.NewSQSClient()
	if err != nil {
		log.Fatalf("Failed to create SQS client: %v", err)
	}

	gqlClient := graphql.NewClient()

	// Initialize repositories
	blockRepo := repository.NewBlockRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	eventRepo := repository.NewEventRepository(db)

	// Initialize services
	eventService := service.NewEventService(eventRepo, sqsClient)
	transactionService := service.NewTransactionService(transactionRepo, gqlClient, eventService)
	blockService := service.NewBlockService(blockRepo, transactionRepo, eventRepo, gqlClient, transactionService)

	// Initialize synchronizer service
	synchronizerService := service.NewSynchronizerService(blockRepo, blockService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := synchronizerService.Start(ctx); err != nil {
			log.Fatalf("Block synchronizer error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down Block Synchronizer...")

	cancel()
	time.Sleep(5 * time.Second)
	log.Println("Block Synchronizer stopped")
}
