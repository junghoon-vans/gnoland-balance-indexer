package repository

import (
	"fmt"
	"os"

	"shared/pkg/database"
	"shared/pkg/models"
)

type ProcessedEventRepository interface {
	IsEventProcessed(eventIdentifier string) (bool, error)
	MarkEventProcessed(event *models.ProcessedEvent) error
}

type processedEventRepository struct {
	db *database.Database
}

func NewProcessedEventRepository(db *database.Database) ProcessedEventRepository {
	return &processedEventRepository{
		db: db,
	}
}

func (r *processedEventRepository) IsEventProcessed(eventIdentifier string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProcessedEvent{}).
		Where("event_identifier = ?", eventIdentifier).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check if event is processed: %w", err)
	}

	return count > 0, nil
}

func (r *processedEventRepository) MarkEventProcessed(event *models.ProcessedEvent) error {
	// Set processor instance if not provided
	if event.ProcessorInstance == "" {
		hostname, _ := os.Hostname()
		event.ProcessorInstance = hostname
	}

	err := r.db.Create(event).Error
	if err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}
