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

	db := initializeDatabase()
	sqsClient := initializeSQSClient()
	gqlClient := graphql.NewClient()

	synchronizerService := initializeServices(db, sqsClient, gqlClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startSynchronizer(ctx, synchronizerService)
	waitForShutdownSignal()
	gracefulShutdown(cancel)
}

func initializeDatabase() *database.Database {
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func initializeSQSClient() *queue.SQSClient {
	sqsClient, err := queue.NewSQSClient()
	if err != nil {
		log.Fatalf("Failed to create SQS client: %v", err)
	}
	return sqsClient
}

func initializeServices(db *database.Database, sqsClient *queue.SQSClient, gqlClient *graphql.Client) service.SynchronizerService {
	// Initialize repositories
	blockRepo := repository.NewBlockRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	eventRepo := repository.NewEventRepository(db)

	// Initialize services
	eventService := service.NewEventService(eventRepo, sqsClient)
	transactionService := service.NewTransactionService(transactionRepo, gqlClient, eventService)
	blockService := service.NewBlockService(blockRepo, transactionRepo, eventRepo, gqlClient, transactionService)

	// Initialize synchronizer service
	return service.NewSynchronizerService(blockRepo, blockService)
}

func startSynchronizer(ctx context.Context, synchronizerService service.SynchronizerService) {
	go func() {
		if err := synchronizerService.Start(ctx); err != nil {
			log.Fatalf("Block synchronizer error: %v", err)
		}
	}()
}

func waitForShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down Block Synchronizer...")
}

func gracefulShutdown(cancel context.CancelFunc) {
	cancel()
	time.Sleep(5 * time.Second)
	log.Println("Block Synchronizer stopped")
}
