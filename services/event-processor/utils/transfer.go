package utils

import "shared/infra/queue"

func GetTransferType(funcName string) string {
	switch funcName {
	case "Mint":
		return queue.TransferTypeMint
	case "Burn":
		return queue.TransferTypeBurn
	case "Transfer":
		return queue.TransferTypeTransfer
	case "":
		return queue.TransferTypeTransfer
	default:
		return "unknown"
	}
}
