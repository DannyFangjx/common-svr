package health

import (
	"context"
	"net/http"

	"common-svr/internal/common/response"

	"github.com/gin-gonic/gin"
)

type DatabasePinger interface {
	PingContext(ctx context.Context) error
}

type Handler struct {
	database DatabasePinger
}

func NewHandler(database DatabasePinger) *Handler {
	return &Handler{database: database}
}

func (h *Handler) Ping(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{"message": "pong"})
}

func (h *Handler) Test(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{"message": "common-svr is running"})
}

func (h *Handler) Health(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Ready(c *gin.Context) {
	if err := h.database.PingContext(c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, response.Envelope{
			Code:    "NOT_READY",
			Message: "database is unavailable",
		})
		return
	}
	response.Success(c, http.StatusOK, gin.H{"status": "ready"})
}
