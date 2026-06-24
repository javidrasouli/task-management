package middleware

import (
	"task-management/internal/utils/logutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		reqLog := log.With(zap.String("request_id", requestID))

		ctx := logutil.WithContext(c.Request.Context(), reqLog)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", requestID)

		start := time.Now()
		c.Next()

		reqLog.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}
