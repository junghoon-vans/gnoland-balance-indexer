package repository

import (
	"shared/pkg/database"
	"shared/pkg/models"
)

type BlockRepository interface {
	GetLastBlock() (*models.Block, error)
	SaveBlock(block *models.Block) error
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
