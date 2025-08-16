package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"block-synchronizer/dto"
	"block-synchronizer/repository"
	"gnoland-balance-indexer/pkg/models"
)

type TransactionService interface {
	ProcessTransactions(ctx context.Context, blockHeight int64) error
	ProcessTransaction(ctx context.Context, gqlTx *dto.GraphQLTransaction) error
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	gqlClient       GraphQLClientInterface
	eventService    EventService
}

func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	gqlClient GraphQLClientInterface,
	eventService EventService,
) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		gqlClient:       gqlClient,
		eventService:    eventService,
	}
}

func (s *transactionService) ProcessTransactions(ctx context.Context, blockHeight int64) error {
	query := `
		query($blockHeight: Int!) {
			getTransactions(where: {block_height: {eq: $blockHeight}}) {
				index
				hash
				success
				block_height
				gas_wanted
				gas_used
				memo
				response {
					events {
						... on GnoEvent {
							type
							func
							pkg_path
							attrs {
								key
								value
							}
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"blockHeight": blockHeight,
	}

	resp, err := s.gqlClient.Query(query, variables)
	if err != nil {
		return err
	}

	var result struct {
		GetTransactions []dto.GraphQLTransaction `json:"getTransactions"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return err
	}

	for _, tx := range result.GetTransactions {
		if err := s.ProcessTransaction(ctx, &tx); err != nil {
			log.Printf("Error processing transaction %s: %v", tx.Hash, err)
			continue
		}
	}

	return nil
}

func (s *transactionService) ProcessTransaction(ctx context.Context, gqlTx *dto.GraphQLTransaction) error {
	tx := &models.Transaction{
		Hash:        gqlTx.Hash,
		Index:       gqlTx.Index,
		BlockHeight: gqlTx.BlockHeight,
		Success:     gqlTx.Success,
		GasWanted:   gqlTx.GasWanted,
		GasUsed:     gqlTx.GasUsed,
		Memo:        gqlTx.Memo,
	}

	if err := s.transactionRepo.SaveTransaction(tx); err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	for _, event := range gqlTx.Response.Events {
		if err := s.eventService.ProcessEvent(ctx, tx.ID, gqlTx, &event); err != nil {
			log.Printf("Error processing event: %v", err)
			continue
		}
	}

	return nil
}