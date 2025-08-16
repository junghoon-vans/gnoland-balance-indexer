package models

import (
	"time"
)

type TokenBalance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Address   string    `gorm:"not null;index" json:"address"`
	TokenPath string    `gorm:"not null;index" json:"token_path"`
	Amount    string    `gorm:"not null" json:"amount"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type TokenTransfer struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	BlockHeight  int64     `gorm:"not null;index" json:"block_height"`
	TxHash       string    `gorm:"not null;index" json:"tx_hash"`
	EventID      uint      `gorm:"not null;index" json:"event_id"`
	FromAddress  string    `gorm:"index" json:"from_address"`
	ToAddress    string    `gorm:"index" json:"to_address"`
	TokenPath    string    `gorm:"not null;index" json:"token_path"`
	Amount       string    `gorm:"not null" json:"amount"`
	TransferType string    `gorm:"not null" json:"transfer_type"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (TokenBalance) TableName() string {
	return "token_balances"
}

func (TokenTransfer) TableName() string {
	return "token_transfers"
}
