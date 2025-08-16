package dto

type GraphQLBlock struct {
	Hash     string `json:"hash"`
	Height   int64  `json:"height"`
	Time     string `json:"time"`
	NumTxs   int    `json:"num_txs"`
	TotalTxs int    `json:"total_txs"`
}