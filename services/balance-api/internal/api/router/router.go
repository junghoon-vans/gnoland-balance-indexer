package router

import (
	"balance-api/internal/api/dto"
	handler2 "balance-api/internal/api/handler"
	middleware2 "balance-api/internal/api/middleware"
	"net/http"
	"strings"

	sharedcache "shared/pkg/cache"

	"github.com/gin-gonic/gin"
)

type Router struct {
	balanceHandler  *handler2.BalanceHandler
	transferHandler *handler2.TransferHandler
	healthHandler   *handler2.HealthHandler
	cache           sharedcache.Cache
}

func NewRouter(
	balanceHandler *handler2.BalanceHandler,
	transferHandler *handler2.TransferHandler,
	healthHandler *handler2.HealthHandler,
	cache sharedcache.Cache,
) *Router {
	return &Router{
		balanceHandler:  balanceHandler,
		transferHandler: transferHandler,
		healthHandler:   healthHandler,
		cache:           cache,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware2.CORSMiddleware())

	// Health check
	router.GET("/health", r.healthHandler.HealthCheck)

	// Balance endpoints with caching
	router.GET("/tokens/balances",
		middleware2.CacheMiddleware(r.cache, middleware2.BalanceAddressConfig),
		r.balanceHandler.GetTokenBalances)

	// Transfer endpoints with caching
	router.GET("/tokens/transfer-history",
		middleware2.CacheMiddleware(r.cache, middleware2.TransferHistoryConfig),
		r.transferHandler.GetTransferHistory)

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
			// Apply cache middleware and call handler
			cacheMiddleware := middleware2.CacheMiddleware(r.cache, middleware2.BalanceTokenConfig)
			cacheMiddleware(c)

			// Only call handler if cache middleware didn't abort
			if !c.IsAborted() {
				r.balanceHandler.GetTokenBalancesByPath(c, tokenPath)
			}
			return
		}
	}
	// If not a token balance request, return 404
	c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Not found"})
}
