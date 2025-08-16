package dto

type GraphQLTransaction struct {
	Index       int                    `json:"index"`
	Hash        string                 `json:"hash"`
	Success     bool                   `json:"success"`
	BlockHeight int64                  `json:"block_height"`
	GasWanted   int64                  `json:"gas_wanted"`
	GasUsed     int64                  `json:"gas_used"`
	Memo        string                 `json:"memo"`
	Response    GraphQLTransactionResp `json:"response"`
}

type GraphQLTransactionResp struct {
	Events []GraphQLEvent `json:"events"`
}