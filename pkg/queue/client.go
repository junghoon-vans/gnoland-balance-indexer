package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSClient() (*SQSClient, error) {
	queueConfig := Load()

	if queueConfig.SQSQueueURL == "" {
		return nil, fmt.Errorf("SQS_QUEUE_URL environment variable is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(queueConfig.AWSRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			queueConfig.AWSAccessKey,
			queueConfig.AWSSecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &SQSClient{
		client:   sqs.NewFromConfig(cfg),
		queueURL: queueConfig.SQSQueueURL,
	}, nil
}

func (s *SQSClient) SendMessage(event *TokenEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(body)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Sending token event to SQS: %s", event.ID)
	_, err = s.client.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	return nil
}

func (s *SQSClient) ReceiveMessages(maxMessages int64) ([]*QueueMessage, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: int32(maxMessages),
		WaitTimeSeconds:     int32(20),
		VisibilityTimeout:   int32(300),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	messages := make([]*QueueMessage, 0, len(result.Messages))
	for _, msg := range result.Messages {
		var event TokenEvent
		if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		queueMsg := &QueueMessage{
			MessageID: *msg.MessageId,
			Body:      event,
		}
		messages = append(messages, queueMsg)
	}

	return messages, nil
}
