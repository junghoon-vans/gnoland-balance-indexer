package service

import (
	"context"
	"errors"
	"testing"

	"shared/pkg/models"
	"shared/pkg/queue"
	"shared/pkg/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBalanceService struct {
	mock.Mock
}

func (m *MockBalanceService) UpdateBalances(ctx context.Context, event *queue.TokenEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockBalanceService) UpdateBalanceAtomic(ctx context.Context, address, tokenPath, amount string) error {
	args := m.Called(ctx, address, tokenPath, amount)
	return args.Error(0)
}

type MockProcessedEventRepository struct {
	mock.Mock
}

func (m *MockProcessedEventRepository) IsEventProcessed(eventIdentifier string) (bool, error) {
	args := m.Called(eventIdentifier)
	return args.Bool(0), args.Error(1)
}

func (m *MockProcessedEventRepository) MarkEventProcessed(event *models.ProcessedEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestNewTokenEventHandler(t *testing.T) {
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}

	handler := NewTokenEventHandler(mockProcessedEventRepo, mockBalanceService)

	assert.NotNil(t, handler)
	assert.IsType(t, &tokenEventHandler{}, handler)
}

func TestTokenEventHandler_ProcessTokenEvent_Success(t *testing.T) {
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockProcessedEventRepo, mockBalanceService)

	event := &queue.TokenEvent{
		ID:          "event1",
		BlockHeight: 100,
		TxHash:      "hash1",
		EventID:     1,
		FromAddress: testutils.TestAddress1,
		ToAddress:   testutils.TestAddress2,
		PkgPath:     testutils.TestTokenPath,
		Amount:      "1000",
		Func:        "Transfer",
		Type:        queue.TransferTypeTransfer,
	}

	// Mock event not processed yet
	mockProcessedEventRepo.On("IsEventProcessed", "hash1-1").Return(false, nil)
	// Mock successful balance update
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(nil)
	// Mock successful event marking
	mockProcessedEventRepo.On("MarkEventProcessed", mock.AnythingOfType("*models.ProcessedEvent")).Return(nil)

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.NoError(t, err)
	mockBalanceService.AssertExpectations(t)
	mockProcessedEventRepo.AssertExpectations(t)
}

func TestTokenEventHandler_ProcessTokenEvent_UpdateBalancesError(t *testing.T) {
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockProcessedEventRepo, mockBalanceService)

	event := &queue.TokenEvent{
		ID:          "event1",
		BlockHeight: 100,
		TxHash:      "hash1",
		EventID:     1,
		FromAddress: testutils.TestAddress1,
		ToAddress:   testutils.TestAddress2,
		PkgPath:     testutils.TestTokenPath,
		Amount:      "1000",
		Func:        "Transfer",
		Type:        queue.TransferTypeTransfer,
	}

	// Mock event not processed yet
	mockProcessedEventRepo.On("IsEventProcessed", "hash1-1").Return(false, nil)
	// Mock balance update error
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(errors.New("balance update error"))

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update balances")
	mockBalanceService.AssertExpectations(t)
}

func TestTokenEventHandler_ProcessTokenEvent_MintEvent(t *testing.T) {
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockProcessedEventRepo, mockBalanceService)

	event := &queue.TokenEvent{
		ID:          "event1",
		BlockHeight: 100,
		TxHash:      "hash1",
		EventID:     1,
		ToAddress:   testutils.TestAddress2,
		PkgPath:     testutils.TestTokenPath,
		Amount:      "1000",
		Func:        "Mint",
		Type:        queue.TransferTypeMint,
	}

	mockProcessedEventRepo.On("IsEventProcessed", "hash1-1").Return(false, nil)
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(nil)
	mockProcessedEventRepo.On("MarkEventProcessed", mock.AnythingOfType("*models.ProcessedEvent")).Return(nil)

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.NoError(t, err)
	mockBalanceService.AssertExpectations(t)
	mockProcessedEventRepo.AssertExpectations(t)
}

// New tests for idempotency logic
func TestTokenEventHandler_ProcessTokenEvent_IdempotentSkip(t *testing.T) {
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockProcessedEventRepo, mockBalanceService)

	event := &queue.TokenEvent{
		ID:          "event1",
		BlockHeight: 100,
		TxHash:      "hash1",
		EventID:     1,
		FromAddress: testutils.TestAddress1,
		ToAddress:   testutils.TestAddress2,
		PkgPath:     testutils.TestTokenPath,
		Amount:      "1000",
		Func:        "Transfer",
		Type:        queue.TransferTypeTransfer,
	}

	// Mock event already processed
	mockProcessedEventRepo.On("IsEventProcessed", "hash1-1").Return(true, nil)

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.NoError(t, err)
	mockProcessedEventRepo.AssertExpectations(t)

	// Verify that balance update was NOT called
	mockBalanceService.AssertNotCalled(t, "UpdateBalances")
}

func TestTokenEventHandler_ProcessTokenEvent_IsEventProcessedError(t *testing.T) {
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockProcessedEventRepo, mockBalanceService)

	event := &queue.TokenEvent{
		ID:          "event1",
		BlockHeight: 100,
		TxHash:      "hash1",
		EventID:     1,
		FromAddress: testutils.TestAddress1,
		ToAddress:   testutils.TestAddress2,
		PkgPath:     testutils.TestTokenPath,
		Amount:      "1000",
		Func:        "Transfer",
		Type:        queue.TransferTypeTransfer,
	}

	// Mock error when checking if event is processed
	mockProcessedEventRepo.On("IsEventProcessed", "hash1-1").Return(false, errors.New("database error"))

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check if event is processed")
	mockProcessedEventRepo.AssertExpectations(t)

	// Verify that balance update was NOT called
	mockBalanceService.AssertNotCalled(t, "UpdateBalances")
}

func TestTokenEventHandler_ProcessTokenEvent_MarkEventProcessedError(t *testing.T) {
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockProcessedEventRepo, mockBalanceService)

	event := &queue.TokenEvent{
		ID:          "event1",
		BlockHeight: 100,
		TxHash:      "hash1",
		EventID:     1,
		FromAddress: testutils.TestAddress1,
		ToAddress:   testutils.TestAddress2,
		PkgPath:     testutils.TestTokenPath,
		Amount:      "1000",
		Func:        "Transfer",
		Type:        queue.TransferTypeTransfer,
	}

	// Mock event not processed yet
	mockProcessedEventRepo.On("IsEventProcessed", "hash1-1").Return(false, nil)
	// Mock successful balance update
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(nil)
	// Mock error when marking event as processed
	mockProcessedEventRepo.On("MarkEventProcessed", mock.AnythingOfType("*models.ProcessedEvent")).Return(errors.New("mark error"))

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	// Should still succeed even if marking failed (graceful degradation)
	assert.NoError(t, err)
	mockBalanceService.AssertExpectations(t)
	mockProcessedEventRepo.AssertExpectations(t)
}
