package service

import (
	"context"
	"testing"

	"shared/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBalanceService struct {
	mock.Mock
}

func (m *MockBalanceService) UpdateBalance(ctx context.Context, address, tokenPath, amount string) error {
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
