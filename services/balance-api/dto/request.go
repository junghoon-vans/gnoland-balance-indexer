package dto

type BalanceRequest struct {
	Address string `form:"address"`
}

type TransferHistoryRequest struct {
	Address string `form:"address"`
	Limit   int    `form:"limit"`
}
