package queue

import (
	"encoding/json"
	"time"
)

// BalanceUpdateMessage represents a single balance update for one address
type BalanceUpdateMessage struct {
	Address      string    `json:"address"`
	TokenPath    string    `json:"token_path"`
	Amount       string    `json:"amount"` // Can be positive or negative
	BlockHeight  int64     `json:"block_height"`
	TxHash       string    `json:"tx_hash"`
	EventID      uint      `json:"event_id"`
	TransferType string    `json:"transfer_type"`
	Timestamp    time.Time `json:"timestamp"`
}

type QueueMessage struct {
	MessageID     string               `json:"message_id"`
	ReceiptHandle string               `json:"receipt_handle"`
	Body          BalanceUpdateMessage `json:"body"`
	Timestamp     time.Time            `json:"timestamp"`
}

func (bum *BalanceUpdateMessage) ToJSON() ([]byte, error) {
	return json.Marshal(bum)
}

func (bum *BalanceUpdateMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, bum)
}

const (
	TransferTypeMint     = "mint"
	TransferTypeBurn     = "burn"
	TransferTypeTransfer = "transfer"
)
