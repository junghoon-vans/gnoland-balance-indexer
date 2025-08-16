package service

import (
	"context"
	"errors"
	"testing"

	"shared/infra/queue"
	"shared/models"
	"shared/testutils"

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

func TestNewTokenEventHandler(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockBalanceService := &MockBalanceService{}

	handler := NewTokenEventHandler(mockTransferRepo, mockBalanceService)

	assert.NotNil(t, handler)
	assert.IsType(t, &tokenEventHandler{}, handler)
}

func TestTokenEventHandler_ProcessTokenEvent_Success(t *testing.T) {
	mockTransferRepo := &MockTransferRepository{}
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockBalanceService)

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

	// Mock successful transfer save
	mockTransferRepo.On("SaveTransfer", mock.AnythingOfType("*models.TokenTransfer")).Return(nil)
	// Mock successful balance update
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(nil)

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.NoError(t, err)
	mockTransferRepo.AssertExpectations(t)
	mockBalanceService.AssertExpectations(t)

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
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockBalanceService)

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
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockBalanceService)

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
	mockBalanceService := &MockBalanceService{}
	handler := NewTokenEventHandler(mockTransferRepo, mockBalanceService)

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

	mockTransferRepo.On("SaveTransfer", mock.AnythingOfType("*models.TokenTransfer")).Return(nil)
	mockBalanceService.On("UpdateBalances", mock.Anything, event).Return(nil)

	err := handler.ProcessTokenEvent(testutils.CreateTestContext(), event)

	assert.NoError(t, err)
	mockTransferRepo.AssertExpectations(t)
	mockBalanceService.AssertExpectations(t)

	// Verify the transfer type is correctly set
	calls := mockTransferRepo.Calls
	transfer := calls[0].Arguments[0].(*models.TokenTransfer)
	assert.Equal(t, queue.TransferTypeMint, transfer.TransferType)
}
