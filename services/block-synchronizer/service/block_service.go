package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"block-synchronizer/dto"
	"block-synchronizer/repository"
	"shared/infra/graphql"
	"shared/models"
)

type GraphQLClientInterface interface {
	Query(query string, variables map[string]interface{}) (*graphql.GraphQLResponse, error)
}

type BlockService interface {
	GetLatestBlockHeight() (int64, error)
	SyncBlockRange(ctx context.Context, startHeight, endHeight int64) error
	ProcessBlock(ctx context.Context, gqlBlock *dto.GraphQLBlock) error
}

type blockService struct {
	blockRepo          repository.BlockRepository
	transactionRepo    repository.TransactionRepository
	eventRepo          repository.EventRepository
	gqlClient          GraphQLClientInterface
	transactionService TransactionService
}

func NewBlockService(
	blockRepo repository.BlockRepository,
	transactionRepo repository.TransactionRepository,
	eventRepo repository.EventRepository,
	gqlClient GraphQLClientInterface,
	transactionService TransactionService,
) BlockService {
	return &blockService{
		blockRepo:          blockRepo,
		transactionRepo:    transactionRepo,
		eventRepo:          eventRepo,
		gqlClient:          gqlClient,
		transactionService: transactionService,
	}
}

func (s *blockService) GetLatestBlockHeight() (int64, error) {
	query := `
		query {
			latestBlockHeight
		}
	`

	resp, err := s.gqlClient.Query(query, nil)
	if err != nil {
		return 0, err
	}

	var result struct {
		LatestBlockHeight int64 `json:"latestBlockHeight"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return 0, err
	}

	return result.LatestBlockHeight, nil
}

func (s *blockService) SyncBlockRange(ctx context.Context, startHeight, endHeight int64) error {
	query := `
		query($startHeight: Int!, $endHeight: Int!) {
			getBlocks(where: {height: {gt: $startHeight, lt: $endHeight}}) {
				hash
				height
				time
				num_txs
				total_txs
			}
		}
	`

	variables := map[string]interface{}{
		"startHeight": startHeight - 1, // gt로 인한 exclusive 처리
		"endHeight":   endHeight + 1,   // lt로 인한 exclusive 처리
	}

	resp, err := s.gqlClient.Query(query, variables)
	if err != nil {
		return err
	}

	var result struct {
		GetBlocks []dto.GraphQLBlock `json:"getBlocks"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return err
	}

	for _, block := range result.GetBlocks {
		if err := s.ProcessBlock(ctx, &block); err != nil {
			log.Printf("Error processing block %d: %v", block.Height, err)
			continue
		}
	}

	return nil
}

func (s *blockService) ProcessBlock(ctx context.Context, gqlBlock *dto.GraphQLBlock) error {
	log.Printf("Processing block %d", gqlBlock.Height)

	parsedTime, err := time.Parse(time.RFC3339, gqlBlock.Time)
	if err != nil {
		return fmt.Errorf("failed to parse block time: %w", err)
	}

	block := &models.Block{
		Hash:     gqlBlock.Hash,
		Height:   gqlBlock.Height,
		Time:     parsedTime,
		NumTxs:   gqlBlock.NumTxs,
		TotalTxs: gqlBlock.TotalTxs,
	}

	if err := s.blockRepo.SaveBlock(block); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	// Process transactions for this block
	log.Printf("Processing transactions for block %d (num_txs: %d)", gqlBlock.Height, gqlBlock.NumTxs)
	if err := s.transactionService.ProcessTransactions(ctx, gqlBlock.Height); err != nil {
		log.Printf("Error processing transactions for block %d: %v", gqlBlock.Height, err)
		// Don't return error here to continue processing other blocks
	} else {
		log.Printf("Successfully processed transactions for block %d", gqlBlock.Height)
	}

	log.Printf("Saved block: %d", block.Height)
	return nil
}
