package middleware

import (
	"context"
	"payment/infrastructure/log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func RequestLogger(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()

		timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
		defer cancel()

		ctx := context.WithValue(timeoutCtx, "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		startTime := time.Now()
		c.Next()
		latency := time.Since(startTime)

		requestLog := logrus.Fields{
			"request_id": requestID,
			"host":       c.Request.Host,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency":    latency,
		}

		if c.Writer.Status() == 200 || c.Writer.Status() == 201 {
			log.Logger.WithFields(requestLog).Info("Request success.")
		} else {
			log.Logger.WithFields(requestLog).Info("Request Error.")
		}
	}
}
