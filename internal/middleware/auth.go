package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware(authService *services.AuthService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger()

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		var blacklistCount int64
		if err := db.Model(&models.TokenBlacklist{}).Where("token = ? AND expires_at > ?", token, time.Now()).Count(&blacklistCount).Error; err == nil && blacklistCount > 0 {
			log.Warn("Blacklisted token used")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Token has been revoked",
			})
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			log.WithError(err).Warn("Invalid token")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			log.Error("Invalid user_id in token claims")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid token claims",
			})
			c.Abort()
			return
		}

		email, ok := claims["email"].(string)
		if !ok {
			log.Error("Invalid email in token claims")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid token claims",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("email", email)

		log.WithField("user_id", userID).Debug("User authenticated successfully")
		c.Next()
	}
}
