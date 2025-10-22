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
	transactionController *controllers.TransactionController,
	dailyTargetController *controllers.DailyTargetController,
	tradingActivityController *controllers.TradingActivityController,
	financialGoalController *controllers.FinancialGoalController,
	backtestStrategyController *controllers.BacktestStrategyController,
	budgetController *controllers.BudgetController,
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

			// Financial Management endpoints
			finance := protected.Group("/finance")
			{
				// Transaction endpoints
				transactions := finance.Group("/transactions")
				{
					transactions.POST("", transactionController.CreateTransaction)
					transactions.GET("", transactionController.GetAllTransactions)
					transactions.GET("/summary", transactionController.GetTransactionSummary)
					transactions.GET("/breakdown", transactionController.GetCategoryBreakdown)
					transactions.GET("/trend", transactionController.GetMonthlyTrend)
					transactions.GET("/category/:category", transactionController.GetTransactionsByCategory)
					transactions.GET("/:id", transactionController.GetTransactionByID)
					transactions.PUT("/:id", transactionController.UpdateTransaction)
					transactions.DELETE("/:id", transactionController.DeleteTransaction)
				}

				// Daily Target endpoints
				targets := finance.Group("/daily-targets")
				{
					targets.POST("", dailyTargetController.CreateDailyTarget)
					targets.GET("", dailyTargetController.GetAllDailyTargets)

					// Specific paths MUST come before wildcard paths
					targets.GET("/today", dailyTargetController.GetTodayTarget)
					targets.GET("/month-summary", dailyTargetController.GetCurrentMonthSummary)
					targets.GET("/week-summary", dailyTargetController.GetWeekSummary)

					// Wildcard paths with nested routes
					targets.POST("/:id/trades", tradingActivityController.AddTradeToTarget)
					targets.GET("/:id/trades", tradingActivityController.GetTradingActivitiesByTarget)
					targets.POST("/:id/refresh", dailyTargetController.RefreshActualValues)

					// Simple wildcard paths at the end
					targets.GET("/:id", dailyTargetController.GetDailyTargetByID)
					targets.PUT("/:id", dailyTargetController.UpdateDailyTarget)
					targets.DELETE("/:id", dailyTargetController.DeleteDailyTarget)
				}

				// Trading Activity endpoints (standalone)
				trading := finance.Group("/trading-activities")
				{
					trading.GET("/stats", tradingActivityController.GetTradingStats)
					trading.GET("/:id", tradingActivityController.GetTradingActivityByID)
					trading.DELETE("/:id", tradingActivityController.DeleteTradingActivity)
				}

				// Financial Goal endpoints
				goals := finance.Group("/goals")
				{
					goals.POST("", financialGoalController.CreateFinancialGoal)
					goals.GET("", financialGoalController.GetAllFinancialGoals)
					goals.GET("/summary", financialGoalController.GetGoalsSummary)
					goals.GET("/:id", financialGoalController.GetFinancialGoalByID)
					goals.PUT("/:id", financialGoalController.UpdateFinancialGoal)
					goals.POST("/:id/update-progress", financialGoalController.UpdateGoalProgress)
					goals.POST("/:id/add-progress", financialGoalController.AddToGoalProgress)
					goals.DELETE("/:id", financialGoalController.DeleteFinancialGoal)
				}

				// Backtest Strategy endpoints
				strategies := finance.Group("/backtest-strategies")
				{
					strategies.POST("", backtestStrategyController.CreateBacktestStrategy)
					strategies.GET("", backtestStrategyController.GetAllBacktestStrategies)
					strategies.GET("/:id", backtestStrategyController.GetBacktestStrategyByID)
					strategies.GET("/:id/result", backtestStrategyController.GetBacktestResult)
					strategies.PUT("/:id", backtestStrategyController.UpdateBacktestStrategy)
					strategies.POST("/:id/run", backtestStrategyController.RunBacktest)
					strategies.POST("/:id/activate", backtestStrategyController.ActivateStrategy)
					strategies.POST("/:id/complete", backtestStrategyController.CompleteStrategy)
					strategies.DELETE("/:id", backtestStrategyController.DeleteBacktestStrategy)
				}

				// Budget endpoints
				budgets := finance.Group("/budgets")
				{
					budgets.POST("", budgetController.CreateBudget)
					budgets.GET("", budgetController.GetAllBudgets)
					budgets.GET("/summary", budgetController.GetBudgetSummary)
					budgets.GET("/monthly-report", budgetController.GetMonthlyBudgetReport)
					budgets.GET("/alerts", budgetController.CheckBudgetAlert)
					budgets.GET("/:id", budgetController.GetBudgetByID)
					budgets.PUT("/:id", budgetController.UpdateBudget)
					budgets.DELETE("/:id", budgetController.DeleteBudget)
				}
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
