package dto

type GraphQLBlock struct {
	Hash     string `json:"hash"`
	Height   int64  `json:"height"`
	Time     string `json:"time"`
	NumTxs   int    `json:"num_txs"`
	TotalTxs int    `json:"total_txs"`
}

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

type GraphQLEvent struct {
	Type    string             `json:"type"`
	Func    string             `json:"func"`
	PkgPath string             `json:"pkg_path"`
	Attrs   []GraphQLEventAttr `json:"attrs"`
}

type GraphQLEventAttr struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}