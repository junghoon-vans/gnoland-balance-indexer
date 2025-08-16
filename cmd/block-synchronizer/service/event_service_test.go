package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"block-synchronizer/dto"
	"gnoland-balance-indexer/pkg/models"
	"gnoland-balance-indexer/pkg/queue"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock repositories for EventService testing
type MockEventRepositoryForService struct {
	mock.Mock
}

func (m *MockEventRepositoryForService) SaveEvent(event *models.TransactionEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventRepositoryForService) SaveEventAttr(attr *models.TransactionEventAttr) error {
	args := m.Called(attr)
	return args.Error(0)
}

type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) SendMessage(ctx context.Context, queueURL string, message interface{}) error {
	args := m.Called(ctx, queueURL, message)
	return args.Error(0)
}

// Test suite for EventService
type EventServiceTestSuite struct {
	suite.Suite
	eventRepo *MockEventRepositoryForService
	sqsClient *MockSQSClient
	service   EventService
}

func (suite *EventServiceTestSuite) SetupTest() {
	suite.eventRepo = new(MockEventRepositoryForService)
	suite.sqsClient = new(MockSQSClient)
	// Create a real SQS client for testing
	realSQSClient := &queue.SQSClient{}
	suite.service = NewEventService(suite.eventRepo, realSQSClient)
}

func (suite *EventServiceTestSuite) TearDownTest() {
	suite.eventRepo.AssertExpectations(suite.T())
	suite.sqsClient.AssertExpectations(suite.T())
}

func (suite *EventServiceTestSuite) TestProcessEvent_Success() {
	ctx := context.Background()
	blockHeight := uint(100)
	tx := &dto.GraphQLTransaction{
		Hash: "test_hash",
	}
	event := &dto.GraphQLEvent{
		Type: "transfer",
		Attrs: []dto.GraphQLEventAttr{
			{
				Key:   "from",
				Value: "addr1",
			},
			{
				Key:   "to",
				Value: "addr2",
			},
			{
				Key:   "amount",
				Value: "1000",
			},
		},
	}

	// Mock expectations
	suite.eventRepo.On("SaveEvent", mock.AnythingOfType("*models.TransactionEvent")).Return(nil)
	suite.eventRepo.On("SaveEventAttr", mock.AnythingOfType("*models.TransactionEventAttr")).Return(nil)

	// Execute
	err := suite.service.ProcessEvent(ctx, blockHeight, tx, event)

	// Assert
	suite.NoError(err)
}

func (suite *EventServiceTestSuite) TestProcessEvent_NoAttrs() {
	ctx := context.Background()
	blockHeight := uint(100)
	tx := &dto.GraphQLTransaction{
		Hash: "test_hash",
	}
	event := &dto.GraphQLEvent{
		Type:  "mint",
		Attrs: []dto.GraphQLEventAttr{},
	}

	// Mock expectations
	suite.eventRepo.On("SaveEvent", mock.AnythingOfType("*models.TransactionEvent")).Return(nil)

	// Execute
	err := suite.service.ProcessEvent(ctx, blockHeight, tx, event)

	// Assert
	suite.NoError(err)
}

func (suite *EventServiceTestSuite) TestProcessEvent_SaveError() {
	ctx := context.Background()
	blockHeight := uint(100)
	tx := &dto.GraphQLTransaction{
		Hash: "test_hash",
	}
	event := &dto.GraphQLEvent{
		Type: "transfer",
		Attrs: []dto.GraphQLEventAttr{
			{
				Key:   "from",
				Value: "addr1",
			},
		},
	}

	// Mock expectations
	suite.eventRepo.On("SaveEvent", mock.AnythingOfType("*models.TransactionEvent")).Return(errors.New("database error"))

	// Execute
	err := suite.service.ProcessEvent(ctx, blockHeight, tx, event)

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "database error")
}

func (suite *EventServiceTestSuite) TestProcessEvent_SaveEventAttrError() {
	ctx := context.Background()
	blockHeight := uint(100)
	tx := &dto.GraphQLTransaction{
		Hash: "test_hash",
	}
	event := &dto.GraphQLEvent{
		Type: "transfer",
		Attrs: []dto.GraphQLEventAttr{
			{
				Key:   "from",
				Value: "addr1",
			},
			{
				Key:   "to",
				Value: "addr2",
			},
		},
	}

	// Mock expectations
	suite.eventRepo.On("SaveEvent", mock.AnythingOfType("*models.TransactionEvent")).Return(nil)
	suite.eventRepo.On("SaveEventAttr", mock.AnythingOfType("*models.TransactionEventAttr")).Return(nil).Once()
	suite.eventRepo.On("SaveEventAttr", mock.AnythingOfType("*models.TransactionEventAttr")).Return(errors.New("attribute save error")).Once()

	// Execute
	err := suite.service.ProcessEvent(ctx, blockHeight, tx, event)

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "attribute save error")
}

func (suite *EventServiceTestSuite) TestProcessEvent_NilEvent() {
	ctx := context.Background()
	blockHeight := uint(100)
	tx := &dto.GraphQLTransaction{
		Hash: "test_hash",
	}

	// Execute
	err := suite.service.ProcessEvent(ctx, blockHeight, tx, nil)

	// Assert
	suite.Error(err)
}

func (suite *EventServiceTestSuite) TestProcessEvent_EmptyType() {
	ctx := context.Background()
	blockHeight := uint(100)
	tx := &dto.GraphQLTransaction{
		Hash: "test_hash",
	}
	event := &dto.GraphQLEvent{
		Type:  "",
		Attrs: []dto.GraphQLEventAttr{},
	}

	// Mock expectations
	suite.eventRepo.On("SaveEvent", mock.AnythingOfType("*models.TransactionEvent")).Return(nil)

	// Execute
	err := suite.service.ProcessEvent(ctx, blockHeight, tx, event)

	// Assert
	suite.NoError(err)
}

func (suite *EventServiceTestSuite) TestProcessEvent_ManyAttrs() {
	ctx := context.Background()
	blockHeight := uint(100)
	tx := &dto.GraphQLTransaction{
		Hash: "test_hash",
	}

	// Create many attributes
	attrs := make([]dto.GraphQLEventAttr, 10)
	for i := 0; i < 10; i++ {
		attrs[i] = dto.GraphQLEventAttr{
			Key:   fmt.Sprintf("key%d", i),
			Value: fmt.Sprintf("value%d", i),
		}
	}

	event := &dto.GraphQLEvent{
		Type:  "complex_event",
		Attrs: attrs,
	}

	// Mock expectations
	suite.eventRepo.On("SaveEvent", mock.AnythingOfType("*models.TransactionEvent")).Return(nil)
	for i := 0; i < 10; i++ {
		suite.eventRepo.On("SaveEventAttr", mock.AnythingOfType("*models.TransactionEventAttr")).Return(nil)
	}

	// Execute
	err := suite.service.ProcessEvent(ctx, blockHeight, tx, event)

	// Assert
	suite.NoError(err)
}

func TestEventServiceTestSuite(t *testing.T) {
	suite.Run(t, new(EventServiceTestSuite))
}