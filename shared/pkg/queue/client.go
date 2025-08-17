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

// SendBalanceUpdate sends a balance update message with MessageGroupId for FIFO ordering
func (s *SQSClient) SendBalanceUpdate(update *BalanceUpdateMessage) error {
	body, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal balance update: %w", err)
	}

	// Create MessageGroupId based on address for ordering
	groupId := update.Address

	// Create MessageDeduplicationId to prevent duplicates
	deduplicationId := fmt.Sprintf("%s-%s-%d", update.TxHash, update.Address, update.EventID)

	input := &sqs.SendMessageInput{
		QueueUrl:               aws.String(s.queueURL),
		MessageBody:            aws.String(string(body)),
		MessageGroupId:         aws.String(groupId),
		MessageDeduplicationId: aws.String(deduplicationId),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Sending balance update to SQS: address=%s, amount=%s", update.Address, update.Amount)
	_, err = s.client.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send balance update to SQS: %w", err)
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
		// Unmarshal as BalanceUpdateMessage
		var balanceUpdate BalanceUpdateMessage
		if err := json.Unmarshal([]byte(*msg.Body), &balanceUpdate); err != nil {
			log.Printf("Failed to unmarshal message as BalanceUpdateMessage: %v", err)
			continue
		}

		queueMsg := &QueueMessage{
			MessageID:     *msg.MessageId,
			ReceiptHandle: *msg.ReceiptHandle,
			Body:          balanceUpdate,
			Timestamp:     time.Now(),
		}
		messages = append(messages, queueMsg)
	}

	return messages, nil
}

func (s *SQSClient) DeleteMessage(receiptHandle string) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.DeleteMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	return nil
}
