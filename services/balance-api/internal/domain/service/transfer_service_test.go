package service

import (
	"balance-api/internal/api/dto"
	"errors"
	"testing"
	"time"

	"shared/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTransferRepository is a mock implementation of TransferRepository
type MockTransferRepository struct {
	mock.Mock
}

func (m *MockTransferRepository) GetTransfersByAddress(address string, limit int) ([]models.TokenTransfer, error) {
	args := m.Called(address, limit)
	return args.Get(0).([]models.TokenTransfer), args.Error(1)
}

func (m *MockTransferRepository) GetAllTransfers(limit int) ([]models.TokenTransfer, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.TokenTransfer), args.Error(1)
}

func TestTransferService_GetTransferHistory_WithAddress(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	// Test data
	now := time.Now()
	transfers := []models.TokenTransfer{
		{
			FromAddress: "g1abcdefghijklmnopqrstuvwxyz1234567890ab",
			ToAddress:   "g1abcdefghijklmnopqrstuvwxyz1234567890cd",
			TokenPath:   "gno.land/r/demo/grc20",
			Amount:      "1000",
			CreatedAt:   now,
		},
		{
			FromAddress: "g1abcdefghijklmnopqrstuvwxyz1234567890cd",
			ToAddress:   "g1abcdefghijklmnopqrstuvwxyz1234567890ab",
			TokenPath:   "gno.land/r/demo/grc21",
			Amount:      "500",
			CreatedAt:   now.Add(-1 * time.Hour),
		},
	}

	mockRepo.On("GetTransfersByAddress", "g1abcdefghijklmnopqrstuvwxyz1234567890ab", 1000).Return(transfers, nil)

	req := dto.TransferHistoryRequest{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", Limit: 0} // 0 should use default limit
	response, err := service.GetTransferHistory(req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Transfers, 2)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890ab", response.Transfers[0].FromAddress)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890cd", response.Transfers[0].ToAddress)
	assert.Equal(t, "gno.land/r/demo/grc20", response.Transfers[0].TokenPath)
	assert.Equal(t, "1000", response.Transfers[0].Amount)

	mockRepo.AssertExpectations(t)
}

func TestTransferService_GetTransferHistory_WithAddressAndLimit(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	// Test data
	transfers := []models.TokenTransfer{
		{
			FromAddress: "g1abcdefghijklmnopqrstuvwxyz1234567890ab",
			ToAddress:   "g1abcdefghijklmnopqrstuvwxyz1234567890cd",
			TokenPath:   "gno.land/r/demo/grc20",
			Amount:      "1000",
			CreatedAt:   time.Now(),
		},
	}

	mockRepo.On("GetTransfersByAddress", "g1abcdefghijklmnopqrstuvwxyz1234567890ab", 10).Return(transfers, nil)

	req := dto.TransferHistoryRequest{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", Limit: 10}
	response, err := service.GetTransferHistory(req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Transfers, 1)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890ab", response.Transfers[0].FromAddress)

	mockRepo.AssertExpectations(t)
}

func TestTransferService_GetTransferHistory_WithInvalidAddress(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	req := dto.TransferHistoryRequest{Address: "invalid_address", Limit: 10}
	response, err := service.GetTransferHistory(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid address format")

	mockRepo.AssertExpectations(t)
}

func TestTransferService_GetTransferHistory_WithoutAddress(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	// Test data
	transfers := []models.TokenTransfer{
		{
			FromAddress: "g1abcdefghijklmnopqrstuvwxyz1234567890ab",
			ToAddress:   "g1abcdefghijklmnopqrstuvwxyz1234567890cd",
			TokenPath:   "gno.land/r/demo/grc20",
			Amount:      "1000",
			CreatedAt:   time.Now(),
		},
		{
			FromAddress: "g1abcdefghijklmnopqrstuvwxyz1234567890ef",
			ToAddress:   "g1abcdefghijklmnopqrstuvwxyz1234567890gh",
			TokenPath:   "gno.land/r/demo/grc21",
			Amount:      "2000",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
	}

	mockRepo.On("GetAllTransfers", 1000).Return(transfers, nil)

	req := dto.TransferHistoryRequest{Address: "", Limit: 0} // Empty address, default limit
	response, err := service.GetTransferHistory(req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Transfers, 2)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890ab", response.Transfers[0].FromAddress)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890ef", response.Transfers[1].FromAddress)

	mockRepo.AssertExpectations(t)
}

func TestTransferService_GetTransferHistory_WithoutAddressAndLimit(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	// Test data
	transfers := []models.TokenTransfer{
		{
			FromAddress: "g1abcdefghijklmnopqrstuvwxyz1234567890ab",
			ToAddress:   "g1abcdefghijklmnopqrstuvwxyz1234567890cd",
			TokenPath:   "gno.land/r/demo/grc20",
			Amount:      "1000",
			CreatedAt:   time.Now(),
		},
	}

	mockRepo.On("GetAllTransfers", 50).Return(transfers, nil)

	req := dto.TransferHistoryRequest{Address: "", Limit: 50}
	response, err := service.GetTransferHistory(req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Transfers, 1)
	assert.Equal(t, "g1abcdefghijklmnopqrstuvwxyz1234567890ab", response.Transfers[0].FromAddress)

	mockRepo.AssertExpectations(t)
}

func TestTransferService_GetTransferHistory_RepositoryError_WithAddress(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	mockRepo.On("GetTransfersByAddress", "g1abcdefghijklmnopqrstuvwxyz1234567890ab", 1000).Return([]models.TokenTransfer{}, errors.New("database error"))

	req := dto.TransferHistoryRequest{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", Limit: 0}
	response, err := service.GetTransferHistory(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to fetch transfer history")

	mockRepo.AssertExpectations(t)
}

func TestTransferService_GetTransferHistory_RepositoryError_WithoutAddress(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	mockRepo.On("GetAllTransfers", 1000).Return([]models.TokenTransfer{}, errors.New("database error"))

	req := dto.TransferHistoryRequest{Address: "", Limit: 0}
	response, err := service.GetTransferHistory(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to fetch transfer history")

	mockRepo.AssertExpectations(t)
}

func TestTransferService_GetTransferHistory_NegativeLimit(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	// Test data
	transfers := []models.TokenTransfer{
		{
			FromAddress: "g1address1",
			ToAddress:   "g1address2",
			TokenPath:   "gno.land/r/demo/grc20",
			Amount:      "1000",
			CreatedAt:   time.Now(),
		},
	}

	// Negative limit should be converted to default limit (1000)
	mockRepo.On("GetTransfersByAddress", "g1abcdefghijklmnopqrstuvwxyz1234567890ab", 1000).Return(transfers, nil)

	req := dto.TransferHistoryRequest{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", Limit: -5}
	response, err := service.GetTransferHistory(req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Transfers, 1)

	mockRepo.AssertExpectations(t)
}

func TestTransferService_GetTransferHistory_EmptyResult(t *testing.T) {
	mockRepo := new(MockTransferRepository)
	service := NewTransferService(mockRepo)

	mockRepo.On("GetTransfersByAddress", "g1abcdefghijklmnopqrstuvwxyz1234567890ab", 1000).Return([]models.TokenTransfer{}, nil)

	req := dto.TransferHistoryRequest{Address: "g1abcdefghijklmnopqrstuvwxyz1234567890ab", Limit: 0}
	response, err := service.GetTransferHistory(req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Transfers, 0)

	mockRepo.AssertExpectations(t)
}
