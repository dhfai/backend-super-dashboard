package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TradingActivityService struct {
	db *gorm.DB
}

func NewTradingActivityService(db *gorm.DB) *TradingActivityService {
	return &TradingActivityService{db: db}
}

// AddTradeToTarget adds a trading activity to a daily target and updates target stats
func (s *TradingActivityService) AddTradeToTarget(userID uuid.UUID, targetID uint, req *models.CreateTradingActivityRequest) (*models.TradingActivity, *models.DailyTarget, error) {
	// Start transaction FIRST
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get daily target INSIDE transaction with lock
	var target models.DailyTarget
	if err := tx.Where("id = ? AND user_id = ?", targetID, userID).First(&target).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("%w: daily target not found", utils.ErrNotFound)
		}
		return nil, nil, err
	}

	// Check if target is already completed
	if target.IsCompleted {
		tx.Rollback()
		return nil, nil, fmt.Errorf("%w: daily target is already completed", utils.ErrInvalidInput)
	}

	// Set trade time to now if not provided
	tradeTime := req.TradeTime
	if tradeTime.IsZero() {
		tradeTime = time.Now()
	}

	// Create trading activity
	activity := &models.TradingActivity{
		DailyTargetID: targetID,
		UserID:        userID,
		TradeType:     req.TradeType,
		Amount:        req.Amount,
		Pips:          req.Pips,
		LotSize:       req.LotSize,
		Symbol:        req.Symbol,
		Description:   req.Description,
		TradeTime:     tradeTime,
	}

	// Save trading activity
	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// Update daily target stats
	target.TotalTrades++

	if req.TradeType == "win" {
		// Add to income (profit)
		target.ActualIncome += req.Amount
		target.WinningTrades++
	} else if req.TradeType == "loss" {
		// Add to expense (loss)
		target.ActualExpense += req.Amount
		target.LosingTrades++
	}

	// Calculate net savings (profit - loss)
	target.ActualSavings = target.ActualIncome - target.ActualExpense

	// Calculate remaining targets
	target.RemainingIncome = target.IncomeTarget - target.ActualIncome
	if target.RemainingIncome < 0 {
		target.RemainingIncome = 0 // Target exceeded
	}

	target.RemainingExpense = target.ExpenseLimit - target.ActualExpense
	if target.RemainingExpense < 0 {
		target.RemainingExpense = 0 // Loss limit exceeded
	}

	// Calculate win rate
	if target.TotalTrades > 0 {
		target.WinRate = (float64(target.WinningTrades) / float64(target.TotalTrades)) * 100
	}

	// Check if income target is reached
	if target.ActualIncome >= target.IncomeTarget && target.IncomeTarget > 0 {
		target.IsCompleted = true
		now := time.Now()
		target.CompletedAt = &now
	}

	// Check if expense limit is exceeded (stop loss)
	if target.ActualExpense >= target.ExpenseLimit && target.ExpenseLimit > 0 {
		target.IsCompleted = true
		now := time.Now()
		target.CompletedAt = &now
	}

	// Save updated target with explicit field selection to ensure all fields are updated
	if err := tx.Model(&target).Updates(map[string]interface{}{
		"total_trades":      target.TotalTrades,
		"winning_trades":    target.WinningTrades,
		"losing_trades":     target.LosingTrades,
		"actual_income":     target.ActualIncome,
		"actual_expense":    target.ActualExpense,
		"actual_savings":    target.ActualSavings,
		"remaining_income":  target.RemainingIncome,
		"remaining_expense": target.RemainingExpense,
		"win_rate":          target.WinRate,
		"is_completed":      target.IsCompleted,
		"completed_at":      target.CompletedAt,
	}).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	return activity, &target, nil
}

// GetTradingActivitiesByTarget retrieves all trading activities for a specific daily target
func (s *TradingActivityService) GetTradingActivitiesByTarget(userID uuid.UUID, targetID uint) ([]models.TradingActivity, error) {
	var activities []models.TradingActivity
	if err := s.db.Where("daily_target_id = ? AND user_id = ?", targetID, userID).
		Order("trade_time desc").
		Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

// GetTradingActivityByID retrieves a specific trading activity
func (s *TradingActivityService) GetTradingActivityByID(userID uuid.UUID, activityID uint) (*models.TradingActivity, error) {
	var activity models.TradingActivity
	if err := s.db.Where("id = ? AND user_id = ?", activityID, userID).First(&activity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: trading activity not found", utils.ErrNotFound)
		}
		return nil, err
	}
	return &activity, nil
}

// DeleteTradingActivity deletes a trading activity and recalculates target stats
func (s *TradingActivityService) DeleteTradingActivity(userID uuid.UUID, activityID uint) error {
	// Start transaction FIRST
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get trading activity INSIDE transaction
	var activity models.TradingActivity
	if err := tx.Where("id = ? AND user_id = ?", activityID, userID).First(&activity).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: trading activity not found", utils.ErrNotFound)
		}
		return err
	}

	// Get daily target INSIDE transaction
	var target models.DailyTarget
	if err := tx.Where("id = ? AND user_id = ?", activity.DailyTargetID, userID).First(&target).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete activity
	if err := tx.Delete(&activity).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Recalculate target stats
	target.TotalTrades--

	if activity.TradeType == "win" {
		target.ActualIncome -= activity.Amount
		target.WinningTrades--
	} else if activity.TradeType == "loss" {
		target.ActualExpense -= activity.Amount
		target.LosingTrades--
	}

	// Recalculate net savings
	target.ActualSavings = target.ActualIncome - target.ActualExpense

	// Recalculate remaining
	target.RemainingIncome = target.IncomeTarget - target.ActualIncome
	if target.RemainingIncome < 0 {
		target.RemainingIncome = 0
	}

	target.RemainingExpense = target.ExpenseLimit - target.ActualExpense
	if target.RemainingExpense < 0 {
		target.RemainingExpense = 0
	}

	// Recalculate win rate
	if target.TotalTrades > 0 {
		target.WinRate = (float64(target.WinningTrades) / float64(target.TotalTrades)) * 100
	} else {
		target.WinRate = 0
	}

	// Reset completion if needed
	if target.IsCompleted {
		// Check if still meets completion criteria
		if target.ActualIncome < target.IncomeTarget && target.ActualExpense < target.ExpenseLimit {
			target.IsCompleted = false
			target.CompletedAt = nil
		}
	}

	// Save updated target with explicit field selection
	if err := tx.Model(&target).Updates(map[string]interface{}{
		"total_trades":      target.TotalTrades,
		"winning_trades":    target.WinningTrades,
		"losing_trades":     target.LosingTrades,
		"actual_income":     target.ActualIncome,
		"actual_expense":    target.ActualExpense,
		"actual_savings":    target.ActualSavings,
		"remaining_income":  target.RemainingIncome,
		"remaining_expense": target.RemainingExpense,
		"win_rate":          target.WinRate,
		"is_completed":      target.IsCompleted,
		"completed_at":      target.CompletedAt,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// GetTradingStats retrieves trading statistics for a date range
func (s *TradingActivityService) GetTradingStats(userID uuid.UUID, dateFrom, dateTo time.Time) (map[string]interface{}, error) {
	var activities []models.TradingActivity

	query := s.db.Where("user_id = ?", userID)
	if !dateFrom.IsZero() {
		query = query.Where("trade_time >= ?", dateFrom)
	}
	if !dateTo.IsZero() {
		query = query.Where("trade_time <= ?", dateTo)
	}

	if err := query.Find(&activities).Error; err != nil {
		return nil, err
	}

	// Calculate stats
	totalTrades := len(activities)
	winningTrades := 0
	losingTrades := 0
	totalProfit := 0.0
	totalLoss := 0.0

	for _, activity := range activities {
		if activity.TradeType == "win" {
			winningTrades++
			totalProfit += activity.Amount
		} else if activity.TradeType == "loss" {
			losingTrades++
			totalLoss += activity.Amount
		}
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = (float64(winningTrades) / float64(totalTrades)) * 100
	}

	netProfit := totalProfit - totalLoss

	stats := map[string]interface{}{
		"total_trades":   totalTrades,
		"winning_trades": winningTrades,
		"losing_trades":  losingTrades,
		"win_rate":       winRate,
		"total_profit":   totalProfit,
		"total_loss":     totalLoss,
		"net_profit":     netProfit,
		"date_from":      dateFrom,
		"date_to":        dateTo,
	}

	return stats, nil
}
