package main

import (
	handler2 "balance-api/internal/api/handler"
	"balance-api/internal/api/router"
	repository2 "balance-api/internal/domain/repository"
	service2 "balance-api/internal/domain/service"
	"log"
	"net/http"
	"os"
	"time"

	"shared/pkg/cache"
	"shared/pkg/database"
)

func main() {
	log.Println("Starting Balance API...")

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	// Initialize repositories
	balanceRepo := repository2.NewBalanceRepository(db)
	transferRepo := repository2.NewTransferRepository(db)

	// Initialize services
	balanceService := service2.NewBalanceService(balanceRepo)
	transferService := service2.NewTransferService(transferRepo)

	// Initialize handlers
	balanceHandler := handler2.NewBalanceHandler(balanceService)
	transferHandler := handler2.NewTransferHandler(transferService)
	healthHandler := handler2.NewHealthHandler()

	// Initialize router
	appRouter := router.NewRouter(balanceHandler, transferHandler, healthHandler, redisCache)
	ginRouter := appRouter.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      ginRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Balance API server starting on port %s", port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
