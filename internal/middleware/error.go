package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log := logger.GetLogger()

		if err, ok := recovered.(string); ok {
			log.WithField("error", err).
				WithField("stack", string(debug.Stack())).
				Error("Panic recovered")
		} else {
			log.WithField("error", recovered).
				WithField("stack", string(debug.Stack())).
				Error("Panic recovered")
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
	})
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
