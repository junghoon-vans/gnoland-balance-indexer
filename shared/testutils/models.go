package testutils

import (
	"time"

	"shared/models"
)

// Block test helpers
func CreateTestBlock(height int64, hash string) *models.Block {
	return &models.Block{
		Hash:     hash,
		Height:   height,
		Time:     time.Now(),
		NumTxs:   5,
		TotalTxs: 100,
	}
}

// Transaction test helpers
func CreateTestTransaction(hash string, blockHeight int64) *models.Transaction {
	return &models.Transaction{
		Hash:        hash,
		Index:       0,
		BlockHeight: blockHeight,
		Success:     true,
		GasWanted:   100000,
		GasUsed:     50000,
		Memo:        "test transaction",
	}
}

// Transaction Event test helpers
func CreateTestTransactionEvent(transactionID uint, eventType string) *models.TransactionEvent {
	return &models.TransactionEvent{
		TransactionID: transactionID,
		Type:          eventType,
		Func:          "Transfer",
		PkgPath:       "gno.land/r/demo/grc20",
	}
}

func CreateTestTransactionEventAttr(eventID uint, key, value string) *models.TransactionEventAttr {
	return &models.TransactionEventAttr{
		EventID: eventID,
		Key:     key,
		Value:   value,
	}
}

// Token Balance test helpers
func CreateTestBalance(address, tokenPath, amount string) *models.TokenBalance {
	return &models.TokenBalance{
		Address:   address,
		TokenPath: tokenPath,
		Amount:    amount,
	}
}

// Token Transfer test helpers
func CreateTestTransfer(blockHeight int64, txHash string, eventID uint, fromAddr, toAddr, tokenPath, amount, transferType string) *models.TokenTransfer {
	return &models.TokenTransfer{
		BlockHeight:  blockHeight,
		TxHash:       txHash,
		EventID:      eventID,
		FromAddress:  fromAddr,
		ToAddress:    toAddr,
		TokenPath:    tokenPath,
		Amount:       amount,
		TransferType: transferType,
	}
}

func CreateTestMintTransfer(blockHeight int64, txHash string, eventID uint, toAddr, tokenPath, amount string) *models.TokenTransfer {
	return CreateTestTransfer(blockHeight, txHash, eventID, "", toAddr, tokenPath, amount, "mint")
}

func CreateTestBurnTransfer(blockHeight int64, txHash string, eventID uint, fromAddr, tokenPath, amount string) *models.TokenTransfer {
	return CreateTestTransfer(blockHeight, txHash, eventID, fromAddr, "", tokenPath, amount, "burn")
}

func CreateTestNormalTransfer(blockHeight int64, txHash string, eventID uint, fromAddr, toAddr, tokenPath, amount string) *models.TokenTransfer {
	return CreateTestTransfer(blockHeight, txHash, eventID, fromAddr, toAddr, tokenPath, amount, "transfer")
}