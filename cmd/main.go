package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/config"
	"github.com/dhfai/dhfai-gobackend.git/internal/controllers"
	"github.com/dhfai/dhfai-gobackend.git/internal/routes"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.InitLogger(cfg.App.Environment)
	log := logger.GetLogger()

	log.Info("Starting File Store API server...")
	log.WithField("environment", cfg.App.Environment).Info("Configuration loaded")

	database, err := config.NewDatabase(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.WithError(err).Fatal("Failed to run database migrations")
	}

	authService := services.NewAuthService(cfg.JWT.Secret)
	emailService := services.NewEmailService(&cfg.Email)
	cleanupService := services.NewCleanupService(database.DB)
	noteService := services.NewNoteService(database.DB)

	// Financial Management Services
	transactionService := services.NewTransactionService(database.DB)
	dailyTargetService := services.NewDailyTargetService(database.DB, transactionService)
	tradingActivityService := services.NewTradingActivityService(database.DB)
	financialGoalService := services.NewFinancialGoalService(database.DB)
	backtestStrategyService := services.NewBacktestStrategyService(database.DB)
	budgetService := services.NewBudgetService(database.DB, transactionService)

	// Run initial cleanup
	cleanupService.RunCleanup()

	// Start periodic cleanup (every 24 hours)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupService.RunCleanup()
		}
	}()

	authController := controllers.NewAuthController(database.DB, authService, emailService)
	profileController := controllers.NewProfileController(database.DB)
	noteController := controllers.NewNoteController(noteService)

	// Financial Management Controllers
	transactionController := controllers.NewTransactionController(transactionService)
	dailyTargetController := controllers.NewDailyTargetController(dailyTargetService)
	tradingActivityController := controllers.NewTradingActivityController(tradingActivityService)
	financialGoalController := controllers.NewFinancialGoalController(financialGoalService)
	backtestStrategyController := controllers.NewBacktestStrategyController(backtestStrategyService)
	budgetController := controllers.NewBudgetController(budgetService)

	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	routes.SetupMiddleware(router)

	routes.SetupRoutes(router, authController, profileController, noteController,
		transactionController, dailyTargetController, tradingActivityController, financialGoalController,
		backtestStrategyController, budgetController, authService, database.DB)

	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	log.WithField("address", serverAddr).Info("Server starting...")

	go func() {
		if err := router.Run(serverAddr); err != nil {
			log.WithError(err).Fatal("Failed to start server")
		}
	}()

	log.WithField("address", serverAddr).Info("Server started successfully")
	log.Info("🚀 File Store API is ready to accept requests!")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	log.Info("Server shutdown complete")
}
