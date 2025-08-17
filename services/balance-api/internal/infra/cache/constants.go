package cache

import (
	"fmt"
	"time"
)

// Cache TTL constants
const (
	BalanceAddressTTL  = 10 * time.Second // Individual address balances
	BalanceTokenTTL    = 15 * time.Second // Token-specific balances
	TransferHistoryTTL = 30 * time.Second // Transfer history
)

// Cache key prefixes
const (
	BalanceAddressPrefix = "balance:address:"
	BalanceTokenPrefix   = "balance:token:"
	TransferPrefix       = "transfer:"
)

func GenerateBalanceAddressKey(address string) string {
	if address != "" {
		return fmt.Sprintf("%s%s", BalanceAddressPrefix, address)
	}
	return fmt.Sprintf("%sall", BalanceAddressPrefix)
}

func GenerateBalanceTokenKey(tokenPath string, address string) string {
	if address != "" {
		return fmt.Sprintf("%s%s:address:%s", BalanceTokenPrefix, tokenPath, address)
	}
	return fmt.Sprintf("%s%s:all", BalanceTokenPrefix, tokenPath)
}

func GenerateTransferKey(address string, limit int) string {
	if address != "" {
		return fmt.Sprintf("%saddress:%s:limit:%d", TransferPrefix, address, limit)
	}
	return fmt.Sprintf("%sall:limit:%d", TransferPrefix, limit)
}
