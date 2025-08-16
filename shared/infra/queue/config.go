package queue

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AWSRegion      string
	AWSAccessKey   string
	AWSSecretKey   string
	SQSQueueURL    string
	AWSEndpointURL string
}

func Load() *Config {
	if err := godotenv.Load("config/.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	return &Config{
		AWSRegion:      getEnv("AWS_REGION", "ap-northeast-2"),
		AWSAccessKey:   getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretKey:   getEnv("AWS_SECRET_ACCESS_KEY", ""),
		SQSQueueURL:    getEnv("SQS_QUEUE_URL", ""),
		AWSEndpointURL: getEnv("AWS_ENDPOINT_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
