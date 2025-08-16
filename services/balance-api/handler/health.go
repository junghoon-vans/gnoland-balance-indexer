package handler

import (
	"net/http"

	"balance-api/dto"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{
		Status:  "healthy",
		Service: "balance-api",
	})
}
