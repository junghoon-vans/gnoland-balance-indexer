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

type MockTransferRepository struct {
	mock.Mock
}

func (m *MockTransferRepository) SaveTransfer(transfer *models.TokenTransfer) error {
	args := m.Called(transfer)
	return args.Error(0)
}

type MockBalanceService struct {
	mock.Mock
}

func (m *MockBalanceService) UpdateBalances(ctx context.Context, event *queue.TokenEvent) error {
	args := m.Called(ctx, event)
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
	mockTransferRepo := &MockTransferRepository{}
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}

	handler := NewTokenEventHandler(mockTransferRepo, mockProcessedEventRepo, mockBalanceService)

	assert.NotNil(t, handler)
	assert.IsType(t, &tokenEventHandler{}, handler)
}

func TestTokenEventHandler_ProcessTokenEvent_Success(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockProcessedEventRepo, mockBalanceService)

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
	// Mock successful transfer save
	mockTransferRepo.On("SaveTransfer", mock.AnythingOfType("*models.TokenTransfer")).Return(nil)
	// Mock successful balance update
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(nil)
	// Mock successful event marking
	mockProcessedEventRepo.On("MarkEventProcessed", mock.AnythingOfType("*models.ProcessedEvent")).Return(nil)

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.NoError(t, err)
	mockTransferRepo.AssertExpectations(t)
	mockBalanceService.AssertExpectations(t)
	mockProcessedEventRepo.AssertExpectations(t)

	// Verify the SaveTransfer was called with correct data
	calls := mockTransferRepo.Calls
	assert.Len(t, calls, 1)
	transfer := calls[0].Arguments[0].(*models.TokenTransfer)
	assert.Equal(t, event.BlockHeight, transfer.BlockHeight)
	assert.Equal(t, event.TxHash, transfer.TxHash)
	assert.Equal(t, event.EventID, transfer.EventID)
	assert.Equal(t, event.FromAddress, transfer.FromAddress)
	assert.Equal(t, event.ToAddress, transfer.ToAddress)
	assert.Equal(t, event.PkgPath, transfer.TokenPath)
	assert.Equal(t, event.Amount, transfer.Amount)
	assert.Equal(t, queue.TransferTypeTransfer, transfer.TransferType)
}

func TestTokenEventHandler_ProcessTokenEvent_SaveTransferError(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockProcessedEventRepo, mockBalanceService)

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
	// Mock transfer save error
	mockTransferRepo.On("SaveTransfer", mock.AnythingOfType("*models.TokenTransfer")).Return(errors.New("database error"))

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save transfer")
	mockTransferRepo.AssertExpectations(t)
	// Balance service should not be called if transfer save fails
	mockBalanceService.AssertNotCalled(t, "UpdateBalances")
}

func TestTokenEventHandler_ProcessTokenEvent_UpdateBalancesError(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockProcessedEventRepo, mockBalanceService)

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
	// Mock successful transfer save
	mockTransferRepo.On("SaveTransfer", mock.AnythingOfType("*models.TokenTransfer")).Return(nil)
	// Mock balance update error
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(errors.New("balance update error"))

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update balances")
	mockTransferRepo.AssertExpectations(t)
	mockBalanceService.AssertExpectations(t)
}

func TestTokenEventHandler_ProcessTokenEvent_MintEvent(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockProcessedEventRepo, mockBalanceService)

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
	mockTransferRepo.On("SaveTransfer", mock.AnythingOfType("*models.TokenTransfer")).Return(nil)
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(nil)
	mockProcessedEventRepo.On("MarkEventProcessed", mock.AnythingOfType("*models.ProcessedEvent")).Return(nil)

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.NoError(t, err)
	mockTransferRepo.AssertExpectations(t)
	mockBalanceService.AssertExpectations(t)
	mockProcessedEventRepo.AssertExpectations(t)

	// Verify the transfer type is correctly set
	calls := mockTransferRepo.Calls
	transfer := calls[0].Arguments[0].(*models.TokenTransfer)
	assert.Equal(t, queue.TransferTypeMint, transfer.TransferType)
}

// New tests for idempotency logic
func TestTokenEventHandler_ProcessTokenEvent_IdempotentSkip(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockProcessedEventRepo, mockBalanceService)

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

	// Verify that transfer save and balance update were NOT called
	mockTransferRepo.AssertNotCalled(t, "SaveTransfer")
	mockBalanceService.AssertNotCalled(t, "UpdateBalances")
}

func TestTokenEventHandler_ProcessTokenEvent_IsEventProcessedError(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockProcessedEventRepo, mockBalanceService)

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

	// Verify that transfer save and balance update were NOT called
	mockTransferRepo.AssertNotCalled(t, "SaveTransfer")
	mockBalanceService.AssertNotCalled(t, "UpdateBalances")
}

func TestTokenEventHandler_ProcessTokenEvent_MarkEventProcessedError(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockProcessedEventRepo := &MockProcessedEventRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockProcessedEventRepo, mockBalanceService)

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
	// Mock successful transfer save and balance update
	mockTransferRepo.On("SaveTransfer", mock.AnythingOfType("*models.TokenTransfer")).Return(nil)
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(nil)
	// Mock error when marking event as processed
	mockProcessedEventRepo.On("MarkEventProcessed", mock.AnythingOfType("*models.ProcessedEvent")).Return(errors.New("mark error"))

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	// Should still succeed even if marking failed (graceful degradation)
	assert.NoError(t, err)
	mockTransferRepo.AssertExpectations(t)
	mockBalanceService.AssertExpectations(t)
	mockProcessedEventRepo.AssertExpectations(t)
}
