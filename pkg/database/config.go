package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPass          string
	DBName          string
	DBSSLMode       string
}

func Load() *Config {
	if err := godotenv.Load("config/.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	return &Config{
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "indexer"),
		DBPass:          getEnv("DB_PASSWORD", "password"),
		DBName:          getEnv("DB_NAME", "gno_indexer"),
		DBSSLMode:       getEnv("DB_SSL_MODE", "disable"),
	}
}

func (c *Config) DatabaseURL() string {
	return "postgres://" + c.DBUser + ":" + c.DBPass + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=" + c.DBSSLMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}