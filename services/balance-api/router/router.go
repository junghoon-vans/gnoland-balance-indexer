package router

import (
	"net/http"
	"strings"

	"balance-api/dto"
	"balance-api/handler"
	"balance-api/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	balanceHandler  *handler.BalanceHandler
	transferHandler *handler.TransferHandler
	healthHandler   *handler.HealthHandler
}

func NewRouter(
	balanceHandler *handler.BalanceHandler,
	transferHandler *handler.TransferHandler,
	healthHandler *handler.HealthHandler,
) *Router {
	return &Router{
		balanceHandler:  balanceHandler,
		transferHandler: transferHandler,
		healthHandler:   healthHandler,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", r.healthHandler.HealthCheck)

	// Balance endpoints
	router.GET("/tokens/balances", r.balanceHandler.GetTokenBalances)

	// Transfer endpoints
	router.GET("/tokens/transfer-history", r.transferHandler.GetTransferHistory)

	// Custom handler for token paths with slashes
	router.NoRoute(r.handleTokenPaths)

	return router
}

func (r *Router) handleTokenPaths(c *gin.Context) {
	path := c.Request.URL.Path

	// Check if this is a token balance request
	if strings.HasPrefix(path, "/tokens/") && strings.HasSuffix(path, "/balances") {
		// Extract token path
		tokenPath := strings.TrimPrefix(path, "/tokens/")
		tokenPath = strings.TrimSuffix(tokenPath, "/balances")

		if tokenPath != "" {
			// Call the balance handler with the extracted token path
			r.balanceHandler.GetTokenBalancesByPath(c, tokenPath)
			return
		}
	}
	// If not a token balance request, return 404
	c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Not found"})
}
