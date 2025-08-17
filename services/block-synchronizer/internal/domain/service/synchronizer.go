package service

import (
	"block-synchronizer/internal/domain/repository"
	"context"
	"fmt"
	"log"
	"time"
)

type SynchronizerService interface {
	Start(ctx context.Context) error
}

type synchronizerService struct {
	blockRepo    repository.BlockRepository
	blockService BlockService
	lastHeight   int64
	syncInterval time.Duration
}

func NewSynchronizerService(
	blockRepo repository.BlockRepository,
	blockService BlockService,
) SynchronizerService {
	return &synchronizerService{
		blockRepo:    blockRepo,
		blockService: blockService,
		syncInterval: 10 * time.Second,
	}
}

func (s *synchronizerService) Start(ctx context.Context) error {
	log.Println("Block Synchronizer started")

	if err := s.initializeLastHeight(); err != nil {
		return fmt.Errorf("failed to initialize last height: %w", err)
	}

	if err := s.backfillMissingBlocks(ctx); err != nil {
		log.Printf("Warning: Failed to backfill missing blocks: %v", err)
	}

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Block Synchronizer context cancelled")
			return nil
		case <-ticker.C:
			if err := s.syncLatestBlocks(ctx); err != nil {
				log.Printf("Error syncing latest blocks: %v", err)
			}
		}
	}
}

func (s *synchronizerService) initializeLastHeight() error {
	lastBlock, err := s.blockRepo.GetLastBlock()
	if err != nil {
		// Start from block height 0 so backfill starts from 1
		s.lastHeight = 0
		log.Printf("No blocks found in database, starting from height 1")
		return nil
	}

	// Start from block height 0 if no previous blocks
	if lastBlock.Height < 1 {
		s.lastHeight = 0
		log.Printf("Starting from height 1")
	} else {
		s.lastHeight = lastBlock.Height
		log.Printf("Last synced block height: %d", s.lastHeight)
	}
	return nil
}

func (s *synchronizerService) backfillMissingBlocks(ctx context.Context) error {
	log.Println("Starting backfill process...")

	latestHeight, err := s.blockService.GetLatestBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get latest block height: %w", err)
	}

	if s.lastHeight >= latestHeight {
		log.Println("No blocks to backfill")
		return nil
	}

	batchSize := int64(100)
	for start := s.lastHeight + 1; start <= latestHeight; start += batchSize {
		end := start + batchSize - 1
		if end > latestHeight {
			end = latestHeight
		}

		log.Printf("Backfilling blocks %d to %d", start, end)

		if err := s.blockService.SyncBlockRange(ctx, start, end); err != nil {
			return fmt.Errorf("failed to sync block range %d-%d: %w", start, end, err)
		}

		s.lastHeight = end
	}

	log.Printf("Backfill completed up to height %d", s.lastHeight)
	return nil
}

func (s *synchronizerService) syncLatestBlocks(ctx context.Context) error {
	latestHeight, err := s.blockService.GetLatestBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get latest block height: %w", err)
	}

	if s.lastHeight >= latestHeight {
		return nil
	}

	log.Printf("Syncing blocks from %d to %d", s.lastHeight+1, latestHeight)

	if err := s.blockService.SyncBlockRange(ctx, s.lastHeight+1, latestHeight); err != nil {
		return fmt.Errorf("failed to sync latest blocks: %w", err)
	}

	s.lastHeight = latestHeight
	return nil
}
