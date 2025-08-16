package utils

import "strings"

func IsValidAddress(addr string) bool {
	if addr == "" {
		return true
	}
	return strings.HasPrefix(addr, "g1") && len(addr) == 40
}

func IsTokenTransferEvent(eventType, function string, attrs map[string]string) bool {
	if eventType != "Transfer" {
		return false
	}

	// For GRC20 transfers, func can be empty, "Mint", "Burn", or "Transfer"
	if function != "" && function != "Mint" && function != "Burn" && function != "Transfer" {
		return false
	}

	requiredKeys := []string{"from", "to", "value"}
	for _, key := range requiredKeys {
		if _, exists := attrs[key]; !exists {
			return false
		}
	}

	return true
}