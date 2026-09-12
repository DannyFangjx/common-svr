package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const TraceIDKey = "trace_id"

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Request-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set(TraceIDKey, traceID)
		c.Header("X-Request-ID", traceID)
		c.Next()
	}
}
