package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"balance-api/handler"
	"balance-api/repository"
	"balance-api/router"
	"balance-api/service"
	"shared/infra/database"
)

func main() {
	log.Println("Starting Balance API...")

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	balanceRepo := repository.NewBalanceRepository(db)
	transferRepo := repository.NewTransferRepository(db)

	// Initialize services
	balanceService := service.NewBalanceService(balanceRepo)
	transferService := service.NewTransferService(transferRepo)

	// Initialize handlers
	balanceHandler := handler.NewBalanceHandler(balanceService)
	transferHandler := handler.NewTransferHandler(transferService)
	healthHandler := handler.NewHealthHandler()

	// Initialize router
	appRouter := router.NewRouter(balanceHandler, transferHandler, healthHandler)
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
