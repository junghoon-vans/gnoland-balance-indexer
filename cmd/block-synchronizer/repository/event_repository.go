package repository

import (
	"gnoland-balance-indexer/pkg/database"
	"gnoland-balance-indexer/pkg/models"
)

type EventRepository interface {
	SaveEvent(event *models.TransactionEvent) error
	SaveEventAttr(attr *models.TransactionEventAttr) error
}

type eventRepository struct {
	db *database.Database
}

func NewEventRepository(db *database.Database) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) SaveEvent(event *models.TransactionEvent) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) SaveEventAttr(attr *models.TransactionEventAttr) error {
	return r.db.Create(attr).Error
}