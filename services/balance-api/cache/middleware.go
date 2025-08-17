package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	sharedcache "shared/infra/cache"
)

// CacheConfig represents cache configuration for an endpoint
type CacheConfig struct {
	TTL          time.Duration
	KeyGenerator func(*gin.Context) string
	ShouldCache  func(*gin.Context) bool
}

// CacheMiddleware creates a caching middleware with the given configuration
func CacheMiddleware(cache sharedcache.Cache, config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip caching if condition not met
		if config.ShouldCache != nil && !config.ShouldCache(c) {
			c.Next()
			return
		}

		// Generate cache key
		cacheKey := config.KeyGenerator(c)
		ctx := context.Background()

		// Try to get from cache
		var cachedData json.RawMessage
		if err := cache.Get(ctx, cacheKey, &cachedData); err == nil {
			log.Printf("Cache hit for key: %s", cacheKey)
			c.Data(200, "application/json", cachedData)
			c.Abort()
			return
		} else if !errors.Is(err, sharedcache.ErrCacheMiss) {
			log.Printf("Cache error for key %s: %v", cacheKey, err)
		}

		// Capture response
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &responseBody{},
		}
		c.Writer = w

		// Execute handler
		c.Next()

		// Cache the response if successful
		if w.Status() == 200 && len(w.body.data) > 0 {
			if err := cache.Set(ctx, cacheKey, json.RawMessage(w.body.data), config.TTL); err != nil {
				log.Printf("Failed to cache response for key %s: %v", cacheKey, err)
			} else {
				log.Printf("Cached response for key: %s", cacheKey)
			}
		}
	}
}

// responseBody captures the response body
type responseBody struct {
	data []byte
}

func (rb *responseBody) Write(p []byte) (int, error) {
	rb.data = append(rb.data, p...)
	return len(p), nil
}

// responseWriter wraps gin.ResponseWriter to capture response
type responseWriter struct {
	gin.ResponseWriter
	body *responseBody
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.body.Write(data)
	return rw.ResponseWriter.Write(data)
}

// Predefined cache configurations for common endpoints
var (
	// BalanceAddressConfig caches balance responses by address
	BalanceAddressConfig = CacheConfig{
		TTL: BalanceAddressTTL,
		KeyGenerator: func(c *gin.Context) string {
			address := c.Query("address")
			return GenerateBalanceAddressKey(address)
		},
		ShouldCache: func(c *gin.Context) bool {
			return c.Request.Method == "GET"
		},
	}

	// BalanceTokenConfig caches token balance responses
	BalanceTokenConfig = CacheConfig{
		TTL: BalanceTokenTTL,
		KeyGenerator: func(c *gin.Context) string {
			// Extract token path from URL path since it's in NoRoute handler
			path := c.Request.URL.Path
			var tokenPath string
			if strings.HasPrefix(path, "/tokens/") && strings.HasSuffix(path, "/balances") {
				tokenPath = strings.TrimPrefix(path, "/tokens/")
				tokenPath = strings.TrimSuffix(tokenPath, "/balances")
			}
			address := c.Query("address")
			return GenerateBalanceTokenKey(tokenPath, address)
		},
		ShouldCache: func(c *gin.Context) bool {
			return c.Request.Method == "GET"
		},
	}

	// TransferHistoryConfig caches transfer history responses
	TransferHistoryConfig = CacheConfig{
		TTL: TransferHistoryTTL,
		KeyGenerator: func(c *gin.Context) string {
			address := c.Query("address")
			limit := 1000 // default limit
			if limitStr := c.Query("limit"); limitStr != "" {
				if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && parsedLimit == 1 {
					// limit parsed successfully
				} else {
					limit = 1000 // fallback to default
				}
			}
			return GenerateTransferKey(address, limit)
		},
		ShouldCache: func(c *gin.Context) bool {
			return c.Request.Method == "GET"
		},
	}
)
