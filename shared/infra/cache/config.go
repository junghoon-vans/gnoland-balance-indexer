package cache

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	DefaultTTL    time.Duration
}

func Load() *Config {
	dbStr := os.Getenv("REDIS_DB")
	if dbStr == "" {
		dbStr = "0"
	}

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	ttlStr := os.Getenv("REDIS_DEFAULT_TTL")
	if ttlStr == "" {
		ttlStr = "300" // 5 minutes default
	}

	ttlSeconds, err := strconv.Atoi(ttlStr)
	if err != nil {
		ttlSeconds = 300
	}

	return &Config{
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       db,
		DefaultTTL:    time.Duration(ttlSeconds) * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
