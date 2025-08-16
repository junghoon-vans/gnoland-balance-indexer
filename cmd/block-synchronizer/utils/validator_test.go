package utils

import "testing"

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "valid address",
			address:  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			expected: true,
		},
		{
			name:     "empty address",
			address:  "",
			expected: true,
		},
		{
			name:     "invalid prefix",
			address:  "a17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			expected: false,
		},
		{
			name:     "too short",
			address:  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat",
			expected: false,
		},
		{
			name:     "too long",
			address:  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7dd",
			expected: false,
		},
		{
			name:     "wrong prefix case",
			address:  "G17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAddress(tt.address)
			if result != tt.expected {
				t.Errorf("IsValidAddress(%s) = %v, expected %v", tt.address, result, tt.expected)
			}
		})
	}
}

func TestIsTokenTransferEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		function  string
		attrs     map[string]string
		expected  bool
	}{
		{
			name:      "valid transfer event",
			eventType: "Transfer",
			function:  "Transfer",
			attrs: map[string]string{
				"from":  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
				"to":    "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"value": "1000",
			},
			expected: true,
		},
		{
			name:      "valid mint event",
			eventType: "Transfer",
			function:  "Mint",
			attrs: map[string]string{
				"from":  "",
				"to":    "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"value": "1000",
			},
			expected: true,
		},
		{
			name:      "valid burn event",
			eventType: "Transfer",
			function:  "Burn",
			attrs: map[string]string{
				"from":  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
				"to":    "",
				"value": "1000",
			},
			expected: true,
		},
		{
			name:      "empty function",
			eventType: "Transfer",
			function:  "",
			attrs: map[string]string{
				"from":  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
				"to":    "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"value": "1000",
			},
			expected: true,
		},
		{
			name:      "wrong event type",
			eventType: "Approval",
			function:  "Transfer",
			attrs: map[string]string{
				"from":  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
				"to":    "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"value": "1000",
			},
			expected: false,
		},
		{
			name:      "invalid function",
			eventType: "Transfer",
			function:  "InvalidFunction",
			attrs: map[string]string{
				"from":  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
				"to":    "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"value": "1000",
			},
			expected: false,
		},
		{
			name:      "missing from attribute",
			eventType: "Transfer",
			function:  "Transfer",
			attrs: map[string]string{
				"to":    "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"value": "1000",
			},
			expected: false,
		},
		{
			name:      "missing to attribute",
			eventType: "Transfer",
			function:  "Transfer",
			attrs: map[string]string{
				"from":  "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
				"value": "1000",
			},
			expected: false,
		},
		{
			name:      "missing value attribute",
			eventType: "Transfer",
			function:  "Transfer",
			attrs: map[string]string{
				"from": "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
				"to":   "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTokenTransferEvent(tt.eventType, tt.function, tt.attrs)
			if result != tt.expected {
				t.Errorf("IsTokenTransferEvent(%s, %s, %v) = %v, expected %v",
					tt.eventType, tt.function, tt.attrs, result, tt.expected)
			}
		})
	}
}