package main

import (
	"context"
	"event-processor/internal/domain/repository"
	"event-processor/internal/domain/service"
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

	db := initializeDatabase()
	sqsClient := initializeSQSClient()

	processorService := initializeServices(db, sqsClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startProcessor(ctx, processorService)
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

func initializeServices(db *database.Database, sqsClient *queue.SQSClient) service.ProcessorService {
	// Initialize repositories
	balanceRepo := repository.NewBalanceRepository(db)
	processedEventRepo := repository.NewProcessedEventRepository(db)

	// Initialize services
	balanceService := service.NewBalanceService(db, balanceRepo)
	tokenEventHandler := service.NewTokenEventHandler(processedEventRepo, balanceService)
	messageProcessor := service.NewMessageProcessor(sqsClient, tokenEventHandler, 10)
	return service.NewProcessorService(messageProcessor)
}

func startProcessor(ctx context.Context, processorService service.ProcessorService) {
	go func() {
		if err := processorService.Start(ctx); err != nil {
			log.Fatalf("Event processor error: %v", err)
		}
	}()
}

func waitForShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down Event Processor...")
}

func gracefulShutdown(cancel context.CancelFunc) {
	cancel()
	time.Sleep(5 * time.Second)
	log.Println("Event Processor stopped")
}
