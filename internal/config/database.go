package config

import (
	"fmt"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(cfg *Config) (*Database, error) {
	log := logger.GetLogger()

	var gormLogLevel gormlogger.LogLevel
	if cfg.App.Environment == "development" {
		gormLogLevel = gormlogger.Info
	} else {
		gormLogLevel = gormlogger.Error
	}

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		log.WithError(err).Error("Failed to connect to database")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.WithError(err).Error("Failed to get underlying sql.DB")
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Info("Successfully connected to database")
	return &Database{DB: db}, nil
}

func (d *Database) Migrate() error {
	log := logger.GetLogger()

	log.Info("Running database migrations...")

	err := d.DB.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.OTP{},
		&models.TempRegistration{},
		&models.TokenBlacklist{},
		&models.Note{},
		&models.Transaction{},
		&models.DailyTarget{},
		&models.TradingActivity{},
		&models.FinancialGoal{},
		&models.BacktestStrategy{},
		&models.Budget{},
	)
	if err != nil {
		log.WithError(err).Error("Failed to run database migrations")
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Info("Database migrations completed successfully")
	return nil
}

func (d *Database) Close() error {
	log := logger.GetLogger()

	sqlDB, err := d.DB.DB()
	if err != nil {
		log.WithError(err).Error("Failed to get underlying sql.DB for closing")
		return err
	}

	if err := sqlDB.Close(); err != nil {
		log.WithError(err).Error("Failed to close database connection")
		return err
	}

	log.Info("Database connection closed successfully")
	return nil
}

func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
