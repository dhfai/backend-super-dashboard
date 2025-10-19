package routes

import (
	"github.com/dhfai/dhfai-gobackend.git/internal/controllers"
	"github.com/dhfai/dhfai-gobackend.git/internal/middleware"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	router *gin.Engine,
	authController *controllers.AuthController,
	profileController *controllers.ProfileController,
	noteController *controllers.NoteController,
	authService *services.AuthService,
	db *gorm.DB,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "file-store-api",
			"version": "1.0.0",
		})
	})

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/verify-registration", authController.VerifyRegistration)
			auth.POST("/resend-registration-otp", authController.ResendRegistrationOTP)
			auth.POST("/login", authController.Login)
			auth.POST("/verify-email", authController.VerifyEmail)
			auth.POST("/resend-verification", authController.ResendVerification)
			auth.POST("/forget-password", authController.ForgetPassword)
			auth.POST("/reset-password", authController.ResetPassword)
		}

		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(authService, db))
		{
			// Auth protected endpoints
			auth := protected.Group("/auth")
			{
				auth.POST("/logout", authController.Logout)
				auth.POST("/request-delete-account", authController.RequestDeleteAccount)
				auth.POST("/delete-account", authController.DeleteAccount)
			}

			profile := protected.Group("/profile")
			{
				profile.GET("", profileController.GetProfile)
				profile.PUT("", profileController.UpdateProfile)
				profile.DELETE("", profileController.DeleteProfile)
			}

			user := protected.Group("/user")
			{
				user.GET("/info", profileController.GetUserInfo)
			}

			// Notes endpoints
			notes := protected.Group("/notes")
			{
				// Get favorite notes (must be before /:id to avoid route conflict)
				notes.GET("/favorites", noteController.GetFavoriteNotes)
				// Search notes
				notes.GET("/search", noteController.SearchNotes)
				// Get notes by tag
				notes.GET("/tags", noteController.GetNotesByTag)
				// Get notes count
				notes.GET("/count", noteController.GetNotesCount)

				// Basic CRUD operations
				notes.POST("", noteController.CreateNote)
				notes.GET("", noteController.GetAllNotes)
				notes.GET("/:id", noteController.GetNote)
				notes.PUT("/:id", noteController.UpdateNote)
				notes.DELETE("/:id", noteController.DeleteNote)

				// Toggle favorite status
				notes.PATCH("/:id/favorite", noteController.ToggleFavorite)
				// Hard delete
				notes.DELETE("/:id/hard-delete", noteController.HardDeleteNote)
			}
		}
	}
}

// SetupMiddleware configures all middleware
func SetupMiddleware(router *gin.Engine) {
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware())
}
