package utils

import "strings"

func IsValidAddress(addr string) bool {
	if addr == "" {
		return false
	}
	return strings.HasPrefix(addr, "g1") && len(addr) == 40
}
