package handler

import (
	"net/http"

	"balance-api/dto"
	"balance-api/service"

	"github.com/gin-gonic/gin"
)

type BalanceHandler struct {
	balanceService service.BalanceService
}

func NewBalanceHandler(balanceService service.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) GetTokenBalances(c *gin.Context) {
	var req dto.BalanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid query parameters"})
		return
	}

	response, err := h.balanceService.GetTokenBalances(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *BalanceHandler) GetTokenBalancesByPath(c *gin.Context, tokenPath string) {
	address := c.Query("address")

	response, err := h.balanceService.GetTokenBalancesByPath(tokenPath, address)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
