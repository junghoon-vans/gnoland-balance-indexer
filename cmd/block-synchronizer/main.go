package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"block-synchronizer/repository"
	"block-synchronizer/service"
	"gnoland-balance-indexer/pkg/database"
	"gnoland-balance-indexer/pkg/graphql"
	"gnoland-balance-indexer/pkg/queue"
)

func main() {
	log.Println("Starting Block Synchronizer...")

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

	gqlClient := graphql.NewClient()

	// Initialize repository
	blockRepo := repository.NewBlockRepository(db)

	// Initialize service
	synchronizerService := service.NewSynchronizerService(blockRepo, sqsClient, gqlClient)

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
