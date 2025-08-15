package repository

import (
	"gnoland-balance-indexer/pkg/database"
	"gnoland-balance-indexer/pkg/models"
)

type BlockRepository interface {
	GetLastBlock() (*models.Block, error)
	SaveBlock(block *models.Block) error
	SaveTransaction(tx *models.Transaction) error
	SaveEvent(event *models.TransactionEvent) error
	SaveEventAttr(attr *models.TransactionEventAttr) error
}

type blockRepository struct {
	db *database.Database
}

func NewBlockRepository(db *database.Database) BlockRepository {
	return &blockRepository{db: db}
}

func (r *blockRepository) GetLastBlock() (*models.Block, error) {
	var block models.Block
	err := r.db.Order("height DESC").First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *blockRepository) SaveBlock(block *models.Block) error {
	return r.db.Create(block).Error
}

func (r *blockRepository) SaveTransaction(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *blockRepository) SaveEvent(event *models.TransactionEvent) error {
	return r.db.Create(event).Error
}

func (r *blockRepository) SaveEventAttr(attr *models.TransactionEventAttr) error {
	return r.db.Create(attr).Error
}