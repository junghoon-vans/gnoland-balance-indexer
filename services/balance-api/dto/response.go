package dto

type BalanceResponse struct {
	Balances []TokenBalanceInfo `json:"balances"`
}

type TokenBalanceInfo struct {
	TokenPath string `json:"tokenPath"`
	Amount    string `json:"amount"`
}

type AccountBalancesResponse struct {
	AccountBalances []AccountBalanceInfo `json:"accountBalances"`
}

type AccountBalanceInfo struct {
	Address   string `json:"address"`
	TokenPath string `json:"tokenPath"`
	Amount    string `json:"amount"`
}

type TransferHistoryResponse struct {
	Transfers []TransferInfo `json:"transfers"`
}

type TransferInfo struct {
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	TokenPath   string `json:"tokenPath"`
	Amount      string `json:"amount"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
