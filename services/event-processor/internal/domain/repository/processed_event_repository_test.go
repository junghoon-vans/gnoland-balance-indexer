package repository

import (
	"testing"

	"shared/pkg/models"
	"shared/pkg/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessedEventRepository_IsEventProcessed(t *testing.T) {
	db := testutils.SetupInMemoryDBOrPanic()
	repo := NewProcessedEventRepository(db)

	t.Run("should return false for unprocessed event", func(t *testing.T) {
		eventIdentifier := "test-tx-hash-123"

		isProcessed, err := repo.IsEventProcessed(eventIdentifier)

		require.NoError(t, err)
		assert.False(t, isProcessed)
	})

	t.Run("should return true for processed event", func(t *testing.T) {
		eventIdentifier := "test-tx-hash-456"

		// First mark the event as processed
		processedEvent := &models.ProcessedEvent{
			EventIdentifier:   eventIdentifier,
			TxHash:            "test-tx-hash",
			EventID:           456,
			BlockHeight:       12345,
			ProcessorInstance: "test-instance",
		}

		err := repo.MarkEventProcessed(processedEvent)
		require.NoError(t, err)

		// Then check if it's marked as processed
		isProcessed, err := repo.IsEventProcessed(eventIdentifier)

		require.NoError(t, err)
		assert.True(t, isProcessed)
	})
}

func TestProcessedEventRepository_MarkEventProcessed(t *testing.T) {
	db := testutils.SetupInMemoryDBOrPanic()
	repo := NewProcessedEventRepository(db)

	t.Run("should successfully mark event as processed", func(t *testing.T) {
		processedEvent := &models.ProcessedEvent{
			EventIdentifier:   "test-tx-789",
			TxHash:            "0x123abc",
			EventID:           789,
			BlockHeight:       54321,
			ProcessorInstance: "processor-1",
		}

		err := repo.MarkEventProcessed(processedEvent)

		require.NoError(t, err)
		assert.NotZero(t, processedEvent.ID)
		assert.NotZero(t, processedEvent.ProcessedAt)
	})

	t.Run("should set processor instance if not provided", func(t *testing.T) {
		processedEvent := &models.ProcessedEvent{
			EventIdentifier: "test-tx-101112",
			TxHash:          "0x456def",
			EventID:         101112,
			BlockHeight:     98765,
			// ProcessorInstance not set
		}

		err := repo.MarkEventProcessed(processedEvent)

		require.NoError(t, err)
		assert.NotEmpty(t, processedEvent.ProcessorInstance)
	})

	t.Run("should fail to mark duplicate event", func(t *testing.T) {
		eventIdentifier := "duplicate-test-event"

		processedEvent1 := &models.ProcessedEvent{
			EventIdentifier: eventIdentifier,
			TxHash:          "0x789ghi",
			EventID:         999,
			BlockHeight:     11111,
		}

		processedEvent2 := &models.ProcessedEvent{
			EventIdentifier: eventIdentifier, // Same identifier
			TxHash:          "0x789ghi",
			EventID:         999,
			BlockHeight:     11111,
		}

		// First should succeed
		err := repo.MarkEventProcessed(processedEvent1)
		require.NoError(t, err)

		// Second should fail due to unique constraint
		err = repo.MarkEventProcessed(processedEvent2)
		assert.Error(t, err)
	})
}

func TestProcessedEventRepository_Integration(t *testing.T) {
	db := testutils.SetupInMemoryDBOrPanic()
	repo := NewProcessedEventRepository(db)

	t.Run("should handle complete workflow", func(t *testing.T) {
		eventIdentifier := "workflow-test-event"

		// 1. Initially not processed
		isProcessed, err := repo.IsEventProcessed(eventIdentifier)
		require.NoError(t, err)
		assert.False(t, isProcessed)

		// 2. Mark as processed
		processedEvent := &models.ProcessedEvent{
			EventIdentifier: eventIdentifier,
			TxHash:          "0xworkflow",
			EventID:         777,
			BlockHeight:     22222,
		}

		err = repo.MarkEventProcessed(processedEvent)
		require.NoError(t, err)

		// 3. Now should be processed
		isProcessed, err = repo.IsEventProcessed(eventIdentifier)
		require.NoError(t, err)
		assert.True(t, isProcessed)
	})
}
