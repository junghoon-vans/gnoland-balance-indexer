package queue

import (
	"encoding/json"
	"time"
)

type TokenEvent struct {
	ID           string                 `json:"id"`
	BlockHeight  int64                  `json:"block_height"`
	TxHash       string                 `json:"tx_hash"`
	EventID      uint                   `json:"event_id"`
	Type         string                 `json:"type"`
	Func         string                 `json:"func"`
	PkgPath      string                 `json:"pkg_path"`
	FromAddress  string                 `json:"from_address"`
	ToAddress    string                 `json:"to_address"`
	Amount       string                 `json:"amount"`
	Timestamp    time.Time              `json:"timestamp"`
	Attributes   map[string]string      `json:"attributes"`
}

type QueueMessage struct {
	MessageID string      `json:"message_id"`
	Body      TokenEvent  `json:"body"`
	Timestamp time.Time   `json:"timestamp"`
}

func (te *TokenEvent) ToJSON() ([]byte, error) {
	return json.Marshal(te)
}

func (te *TokenEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, te)
}

const (
	TransferTypeMint     = "mint"
	TransferTypeBurn     = "burn"
	TransferTypeTransfer = "transfer"
)