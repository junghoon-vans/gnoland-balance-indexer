package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"shared/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMessageProcessor struct {
	mock.Mock
}

func (m *MockMessageProcessor) ProcessMessages(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewProcessorService(t *testing.T) {
	mockMsgProcessor := &MockMessageProcessor{}

	service := NewProcessorService(mockMsgProcessor)

	assert.NotNil(t, service)
	assert.IsType(t, &processorService{}, service)
}

func TestProcessorService_Start_Success(t *testing.T) {
	mockMsgProcessor := &MockMessageProcessor{}
	service := NewProcessorService(mockMsgProcessor)

	// Create a test context that will be cancelled after enough time for at least one tick
	ctx, cancel := context.WithTimeout(testutils.CreateTestContext(), 6*time.Second)
	defer cancel()

	// Mock successful message processing
	mockMsgProcessor.On("ProcessMessages", mock.Anything).Return(nil)

	err := service.Start(ctx)

	// Should return context.DeadlineExceeded when context times out
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)

	// Verify ProcessMessages was called at least once
	mockMsgProcessor.AssertCalled(t, "ProcessMessages", mock.Anything)
}

func TestProcessorService_Start_ProcessingError(t *testing.T) {
	mockMsgProcessor := &MockMessageProcessor{}
	service := NewProcessorService(mockMsgProcessor)

	// Create a test context that will be cancelled after enough time for at least one tick
	ctx, cancel := context.WithTimeout(testutils.CreateTestContext(), 6*time.Second)
	defer cancel()

	// Mock message processing error
	mockMsgProcessor.On("ProcessMessages", mock.Anything).Return(errors.New("processing error"))

	err := service.Start(ctx)

	// Should return the processing error or context timeout
	assert.Error(t, err)
	// Could be either the processing error or context timeout depending on timing

	mockMsgProcessor.AssertCalled(t, "ProcessMessages", mock.Anything)
}

func TestProcessorService_Start_ContextCancellation(t *testing.T) {
	mockMsgProcessor := &MockMessageProcessor{}
	service := NewProcessorService(mockMsgProcessor)

	// Create a test context that will be cancelled immediately
	ctx, cancel := context.WithCancel(testutils.CreateTestContext())
	cancel() // Cancel immediately

	err := service.Start(ctx)

	// Should return context.Canceled
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}