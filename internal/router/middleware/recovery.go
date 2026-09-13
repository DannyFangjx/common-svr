package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"common-svr/internal/common/response"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.ErrorContext(c.Request.Context(), "panic recovered",
			"panic", fmt.Sprint(recovered),
			"request_id", c.GetString(RequestIDKey),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Envelope{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		})
	})
}
