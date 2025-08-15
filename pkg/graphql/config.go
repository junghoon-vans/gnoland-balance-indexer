package graphql

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GraphQLEndpoint string
}

func Load() *Config {
	if err := godotenv.Load("config/.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	return &Config{
		GraphQLEndpoint: getEnv("GRAPHQL_ENDPOINT", "https://dev-indexer.api.gnoswap.io/graphql/query"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}