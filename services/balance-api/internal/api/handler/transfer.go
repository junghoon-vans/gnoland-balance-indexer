package handler

import (
	"net/http"

	"balance-api/internal/api/dto"
	"balance-api/internal/domain/service"
	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	transferService service.TransferService
}

func NewTransferHandler(transferService service.TransferService) *TransferHandler {
	return &TransferHandler{transferService: transferService}
}

func (h *TransferHandler) GetTransferHistory(c *gin.Context) {
	var req dto.TransferHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid query parameters"})
		return
	}

	response, err := h.transferService.GetTransferHistory(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
