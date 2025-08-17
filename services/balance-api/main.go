package main

import (
	"balance-api/internal/api/handler"
	"balance-api/internal/api/router"
	"balance-api/internal/domain/repository"
	"balance-api/internal/domain/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"shared/pkg/cache"
	"shared/pkg/database"
)

func main() {
	log.Println("Starting Balance API...")

	ginEngine := initializeEngine()
	port := getPort()
	srv := createServer(ginEngine, port)

	startServer(srv, port)
	waitForShutdownSignal()
	gracefulShutdown(srv)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func createServer(ginEngine *gin.Engine, port string) *http.Server {
	return &http.Server{
		Addr:    ":" + port,
		Handler: ginEngine,
	}
}

func startServer(srv *http.Server, port string) {
	go func() {
		log.Printf("Balance API server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
}

func waitForShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down Balance API...")
}

func gracefulShutdown(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Balance API stopped gracefully")
	}
}

func initializeEngine() *gin.Engine {
	// Database connection
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Redis cache connection
	redisCache, err := cache.NewRedisCache()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
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

	// Setup router
	r := router.NewRouter(balanceHandler, transferHandler, healthHandler, redisCache)
	return r.SetupRoutes()
}
