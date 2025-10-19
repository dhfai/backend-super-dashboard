package middleware

import (
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware logs HTTP requests with beautiful formatting
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log := logger.GetLogger()

		var logLevel logrus.Level
		switch {
		case param.StatusCode >= 500:
			logLevel = logrus.ErrorLevel
		case param.StatusCode >= 400:
			logLevel = logrus.WarnLevel
		default:
			logLevel = logrus.InfoLevel
		}

		entry := log.WithFields(logrus.Fields{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency,
			"client_ip":  param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"request_id": param.Request.Header.Get("X-Request-ID"),
		})

		// Log the request
		message := "HTTP Request"
		switch logLevel {
		case logrus.ErrorLevel:
			entry.Error(message)
		case logrus.WarnLevel:
			entry.Warn(message)
		default:
			entry.Info(message)
		}

		return ""
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return uuid.New().String()
}
