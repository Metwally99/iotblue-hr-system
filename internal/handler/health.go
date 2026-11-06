package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatabasePinger interface {
	Ping(context.Context) error
}

type HealthHandler struct {
	database DatabasePinger
}

func NewHealthHandler(database DatabasePinger) *HealthHandler {
	return &HealthHandler{database: database}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok"}})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	if err := h.database.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{"code": "NOT_READY", "message": "database is unavailable", "details": nil},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ready"}})
}
