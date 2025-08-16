package models

import (
	"time"
)

type Block struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Hash      string    `gorm:"unique;not null" json:"hash"`
	Height    int64     `gorm:"unique;not null" json:"height"`
	Time      time.Time `gorm:"not null" json:"time"`
	NumTxs    int       `json:"num_txs"`
	TotalTxs  int       `json:"total_txs"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Transaction struct {
	ID          uint               `gorm:"primaryKey" json:"id"`
	Hash        string             `gorm:"unique;not null" json:"hash"`
	Index       int                `gorm:"not null" json:"index"`
	BlockHeight int64              `gorm:"not null;index" json:"block_height"`
	Success     bool               `gorm:"not null" json:"success"`
	GasWanted   int64              `json:"gas_wanted"`
	GasUsed     int64              `json:"gas_used"`
	Memo        string             `json:"memo"`
	Messages    []TransactionMsg   `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE" json:"messages"`
	Events      []TransactionEvent `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE" json:"events"`
	CreatedAt   time.Time          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
}

type TransactionMsg struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	TransactionID uint   `gorm:"not null;index" json:"transaction_id"`
	Route         string `json:"route"`
	TypeURL       string `json:"type_url"`
	Value         string `gorm:"type:jsonb" json:"value"`
}

type TransactionEvent struct {
	ID            uint                   `gorm:"primaryKey" json:"id"`
	TransactionID uint                   `gorm:"not null;index" json:"transaction_id"`
	Type          string                 `gorm:"not null" json:"type"`
	Func          string                 `json:"func"`
	PkgPath       string                 `json:"pkg_path"`
	Attrs         []TransactionEventAttr `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"attrs"`
}

type TransactionEventAttr struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	EventID uint   `gorm:"not null;index" json:"event_id"`
	Key     string `gorm:"not null" json:"key"`
	Value   string `json:"value"`
}
