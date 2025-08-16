package utils

import (
	"testing"

	"shared/infra/queue"
	"github.com/stretchr/testify/assert"
)

func TestGetTransferType(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		expected string
	}{
		{
			name:     "Mint function",
			funcName: "Mint",
			expected: queue.TransferTypeMint,
		},
		{
			name:     "Burn function",
			funcName: "Burn",
			expected: queue.TransferTypeBurn,
		},
		{
			name:     "Transfer function",
			funcName: "Transfer",
			expected: queue.TransferTypeTransfer,
		},
		{
			name:     "Empty function name",
			funcName: "",
			expected: queue.TransferTypeTransfer,
		},
		{
			name:     "Unknown function",
			funcName: "UnknownFunc",
			expected: "unknown",
		},
		{
			name:     "Case sensitive test",
			funcName: "mint",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTransferType(tt.funcName)
			assert.Equal(t, tt.expected, result)
		})
	}
}