package queue

import (
	"encoding/json"
	"time"
)

type TokenEvent struct {
	ID          string            `json:"id"`
	BlockHeight int64             `json:"block_height"`
	TxHash      string            `json:"tx_hash"`
	EventID     uint              `json:"event_id"`
	Type        string            `json:"type"`
	Func        string            `json:"func"`
	PkgPath     string            `json:"pkg_path"`
	FromAddress string            `json:"from_address"`
	ToAddress   string            `json:"to_address"`
	Amount      string            `json:"amount"`
	Timestamp   time.Time         `json:"timestamp"`
	Attributes  map[string]string `json:"attributes"`
}

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
	MessageID     string      `json:"message_id"`
	ReceiptHandle string      `json:"receipt_handle"`
	Body          interface{} `json:"body"` // Can be TokenEvent or BalanceUpdateMessage
	Timestamp     time.Time   `json:"timestamp"`
}

func (te *TokenEvent) ToJSON() ([]byte, error) {
	return json.Marshal(te)
}

func (te *TokenEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, te)
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
