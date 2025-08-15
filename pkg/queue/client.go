package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gnoland-balance-indexer/pkg/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSClient struct {
	client   *sqs.SQS
	queueURL string
}

func NewSQSClient() (*SQSClient, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-2"
	}

	endpointURL := os.Getenv("AWS_ENDPOINT_URL")
	queueURL := os.Getenv("SQS_QUEUE_URL")

	if queueURL == "" {
		return nil, fmt.Errorf("SQS_QUEUE_URL environment variable is required")
	}

	config := &aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	}

	if endpointURL != "" {
		config.Endpoint = aws.String(endpointURL)
		config.DisableSSL = aws.Bool(true)
		config.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &SQSClient{
		client:   sqs.New(sess),
		queueURL: queueURL,
	}, nil
}

func (s *SQSClient) SendMessage(event *models.TokenEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(body)),
	}

	_, err = s.client.SendMessage(input)
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	return nil
}

func (s *SQSClient) ReceiveMessages(maxMessages int64) ([]*models.QueueMessage, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: aws.Int64(maxMessages),
		WaitTimeSeconds:     aws.Int64(20),
		VisibilityTimeout:   aws.Int64(300),
	}

	result, err := s.client.ReceiveMessage(input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	messages := make([]*models.QueueMessage, 0, len(result.Messages))
	for _, msg := range result.Messages {
		var event models.TokenEvent
		if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		queueMsg := &models.QueueMessage{
			MessageID: *msg.MessageId,
			Body:      event,
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

	_, err := s.client.DeleteMessage(input)
	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	return nil
}