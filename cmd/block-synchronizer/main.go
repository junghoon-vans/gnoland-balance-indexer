package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"block-synchronizer/config"
	"block-synchronizer/repository"
	"block-synchronizer/service"
	"gnoland-balance-indexer/pkg/database"
	"gnoland-balance-indexer/pkg/queue"
)

func main() {
	log.Println("Starting Block Synchronizer...")

	cfg := config.Load()

	// Set DATABASE_URL for pkg/database compatibility
	os.Setenv("DATABASE_URL", cfg.DatabaseURL())

	// Set SQS environment variables for pkg/queue compatibility
	os.Setenv("AWS_REGION", cfg.AWSRegion)
	os.Setenv("AWS_ACCESS_KEY_ID", cfg.AWSAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", cfg.AWSSecretKey)
	os.Setenv("SQS_QUEUE_URL", cfg.SQSQueueURL)
	if cfg.AWSEndpointURL != "" {
		os.Setenv("AWS_ENDPOINT_URL", cfg.AWSEndpointURL)
	}

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := db.CreateUniqueIndexes(); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	sqsClient, err := queue.NewSQSClient()
	if err != nil {
		log.Fatalf("Failed to create SQS client: %v", err)
	}

	// Initialize repository
	blockRepo := repository.NewBlockRepository(db)

	// Initialize service
	synchronizerService := service.NewSynchronizerService(blockRepo, sqsClient, cfg)

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
