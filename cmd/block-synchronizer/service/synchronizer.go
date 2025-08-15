package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"block-synchronizer/dto"
	"block-synchronizer/repository"
	"block-synchronizer/utils"
	"gnoland-balance-indexer/pkg/graphql"
	"gnoland-balance-indexer/pkg/models"
	"gnoland-balance-indexer/pkg/queue"
)

type SynchronizerService interface {
	Start(ctx context.Context) error
}

type synchronizerService struct {
	blockRepo    repository.BlockRepository
	sqsClient    *queue.SQSClient
	gqlClient    *graphql.Client
	lastHeight   int64
	syncInterval time.Duration
}

func NewSynchronizerService(
	blockRepo repository.BlockRepository,
	sqsClient *queue.SQSClient,
	gqlClient *graphql.Client,
) SynchronizerService {
	return &synchronizerService{
		blockRepo:    blockRepo,
		sqsClient:    sqsClient,
		gqlClient:    gqlClient,
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
		s.lastHeight = 0
		log.Println("No blocks found in database, starting from height 0")
		return nil
	}

	s.lastHeight = lastBlock.Height
	log.Printf("Last synced block height: %d", s.lastHeight)
	return nil
}

func (s *synchronizerService) backfillMissingBlocks(ctx context.Context) error {
	log.Println("Starting backfill process...")

	latestHeight, err := s.getLatestBlockHeight()
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

		if err := s.syncBlockRange(ctx, start, end); err != nil {
			return fmt.Errorf("failed to sync block range %d-%d: %w", start, end, err)
		}

		s.lastHeight = end
	}

	log.Printf("Backfill completed up to height %d", s.lastHeight)
	return nil
}

func (s *synchronizerService) syncLatestBlocks(ctx context.Context) error {
	latestHeight, err := s.getLatestBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get latest block height: %w", err)
	}

	if s.lastHeight >= latestHeight {
		return nil
	}

	log.Printf("Syncing blocks from %d to %d", s.lastHeight+1, latestHeight)

	if err := s.syncBlockRange(ctx, s.lastHeight+1, latestHeight); err != nil {
		return fmt.Errorf("failed to sync latest blocks: %w", err)
	}

	s.lastHeight = latestHeight
	return nil
}

func (s *synchronizerService) getLatestBlockHeight() (int64, error) {
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

func (s *synchronizerService) syncBlockRange(ctx context.Context, startHeight, endHeight int64) error {
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
		if err := s.processBlock(ctx, &block); err != nil {
			log.Printf("Error processing block %d: %v", block.Height, err)
			continue
		}
	}

	return nil
}

func (s *synchronizerService) processBlock(ctx context.Context, gqlBlock *dto.GraphQLBlock) error {
	blockTime, err := time.Parse("2006-01-02T15:04:05.999999999Z", gqlBlock.Time)
	if err != nil {
		blockTime = time.Now()
	}

	block := &models.Block{
		Hash:     gqlBlock.Hash,
		Height:   gqlBlock.Height,
		Time:     blockTime,
		NumTxs:   gqlBlock.NumTxs,
		TotalTxs: gqlBlock.TotalTxs,
	}

	if err := s.blockRepo.SaveBlock(block); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	if gqlBlock.NumTxs > 0 {
		return s.processTransactions(ctx, gqlBlock.Height)
	}

	return nil
}

func (s *synchronizerService) processTransactions(ctx context.Context, blockHeight int64) error {
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
		if err := s.processTransaction(ctx, &tx); err != nil {
			log.Printf("Error processing transaction %s: %v", tx.Hash, err)
			continue
		}
	}

	return nil
}

func (s *synchronizerService) processTransaction(ctx context.Context, gqlTx *dto.GraphQLTransaction) error {
	tx := &models.Transaction{
		Hash:        gqlTx.Hash,
		Index:       gqlTx.Index,
		BlockHeight: gqlTx.BlockHeight,
		Success:     gqlTx.Success,
		GasWanted:   gqlTx.GasWanted,
		GasUsed:     gqlTx.GasUsed,
		Memo:        gqlTx.Memo,
	}

	if err := s.blockRepo.SaveTransaction(tx); err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	for _, event := range gqlTx.Response.Events {
		if err := s.processEvent(ctx, tx.ID, gqlTx, &event); err != nil {
			log.Printf("Error processing event: %v", err)
			continue
		}
	}

	return nil
}

func (s *synchronizerService) processEvent(ctx context.Context, txID uint, gqlTx *dto.GraphQLTransaction, gqlEvent *dto.GraphQLEvent) error {
	event := &models.TransactionEvent{
		TransactionID: txID,
		Type:          gqlEvent.Type,
		Func:          gqlEvent.Func,
		PkgPath:       gqlEvent.PkgPath,
	}

	if err := s.blockRepo.SaveEvent(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	attrs := make(map[string]string)
	for _, attr := range gqlEvent.Attrs {
		eventAttr := &models.TransactionEventAttr{
			EventID: event.ID,
			Key:     attr.Key,
			Value:   attr.Value,
		}

		if err := s.blockRepo.SaveEventAttr(eventAttr); err != nil {
			return fmt.Errorf("failed to save event attribute: %w", err)
		}

		attrs[attr.Key] = attr.Value
	}

	if utils.IsTokenTransferEvent(gqlEvent.Type, gqlEvent.Func, attrs) {
		return s.sendTokenEventToQueue(ctx, gqlTx, event, attrs)
	}

	return nil
}

func (s *synchronizerService) sendTokenEventToQueue(ctx context.Context, gqlTx *dto.GraphQLTransaction, event *models.TransactionEvent, attrs map[string]string) error {
	fromAddr := attrs["from"]
	toAddr := attrs["to"]
	amount := attrs["value"]

	if _, err := strconv.ParseInt(amount, 10, 64); err != nil {
		return fmt.Errorf("invalid amount format: %s", amount)
	}

	if !utils.IsValidAddress(fromAddr) && fromAddr != "" {
		return fmt.Errorf("invalid from address: %s", fromAddr)
	}

	if !utils.IsValidAddress(toAddr) && toAddr != "" {
		return fmt.Errorf("invalid to address: %s", toAddr)
	}

	tokenEvent := &models.TokenEvent{
		ID:          fmt.Sprintf("%s-%d", gqlTx.Hash, event.ID),
		BlockHeight: gqlTx.BlockHeight,
		TxHash:      gqlTx.Hash,
		EventID:     event.ID,
		Type:        event.Type,
		Func:        event.Func,
		PkgPath:     event.PkgPath,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      amount,
		Timestamp:   time.Now(),
		Attributes:  attrs,
	}

	return s.sqsClient.SendMessage(tokenEvent)
}