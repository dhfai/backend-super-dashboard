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

type DailyTargetService struct {
	db                 *gorm.DB
	transactionService *TransactionService
}

func NewDailyTargetService(db *gorm.DB, transactionService *TransactionService) *DailyTargetService {
	return &DailyTargetService{
		db:                 db,
		transactionService: transactionService,
	}
}

// CreateDailyTarget creates a new daily target
func (s *DailyTargetService) CreateDailyTarget(userID uuid.UUID, req *models.CreateDailyTargetRequest) (*models.DailyTarget, error) {
	// Check if target already exists for this date
	var existing models.DailyTarget
	if err := s.db.Where("user_id = ? AND date = ?", userID, req.Date).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("%w: daily target already exists for this date", utils.ErrAlreadyExists)
	}

	target := &models.DailyTarget{
		UserID:        userID,
		Date:          req.Date,
		IncomeTarget:  req.IncomeTarget,
		ExpenseLimit:  req.ExpenseLimit,
		SavingsTarget: req.SavingsTarget,
		Notes:         req.Notes,
	}

	// Calculate actual values from transactions
	if err := s.updateActualValues(target); err != nil {
		return nil, err
	}

	if err := s.db.Create(target).Error; err != nil {
		return nil, err
	}

	return target, nil
}

// GetDailyTargetByID retrieves a daily target by ID
func (s *DailyTargetService) GetDailyTargetByID(userID uuid.UUID, targetID uint) (*models.DailyTarget, error) {
	var target models.DailyTarget
	if err := s.db.Where("id = ? AND user_id = ?", targetID, userID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: daily target not found", utils.ErrNotFound)
		}
		return nil, err
	}
	return &target, nil
}

// GetDailyTargetByDate retrieves a daily target by date
func (s *DailyTargetService) GetDailyTargetByDate(userID uuid.UUID, date time.Time) (*models.DailyTarget, error) {
	var target models.DailyTarget
	if err := s.db.Where("user_id = ? AND date = ?", userID, date).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("daily target not found")
		}
		return nil, err
	}
	return &target, nil
}

// GetAllDailyTargets retrieves all daily targets with pagination
func (s *DailyTargetService) GetAllDailyTargets(userID uuid.UUID, page, pageSize int, dateFrom, dateTo *time.Time) ([]models.DailyTarget, int64, error) {
	var targets []models.DailyTarget
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	query := s.db.Model(&models.DailyTarget{}).Where("user_id = ?", userID)

	// Apply date filters
	if dateFrom != nil {
		query = query.Where("date >= ?", dateFrom)
	}
	if dateTo != nil {
		query = query.Where("date <= ?", dateTo)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Order("date DESC").Limit(pageSize).Offset(offset).Find(&targets).Error; err != nil {
		return nil, 0, err
	}

	return targets, total, nil
}

// UpdateDailyTarget updates a daily target
func (s *DailyTargetService) UpdateDailyTarget(userID uuid.UUID, targetID uint, req *models.UpdateDailyTargetRequest) (*models.DailyTarget, error) {
	target, err := s.GetDailyTargetByID(userID, targetID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.IncomeTarget != nil {
		target.IncomeTarget = *req.IncomeTarget
	}
	if req.ExpenseLimit != nil {
		target.ExpenseLimit = *req.ExpenseLimit
	}
	if req.SavingsTarget != nil {
		target.SavingsTarget = *req.SavingsTarget
	}
	if req.Notes != "" {
		target.Notes = req.Notes
	}

	// Update actual values
	if err := s.updateActualValues(target); err != nil {
		return nil, err
	}

	if err := s.db.Save(target).Error; err != nil {
		return nil, err
	}

	return target, nil
}

// DeleteDailyTarget soft deletes a daily target
func (s *DailyTargetService) DeleteDailyTarget(userID uuid.UUID, targetID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", targetID, userID).Delete(&models.DailyTarget{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("daily target not found")
	}
	return nil
}

// RefreshActualValues updates actual values from transactions
func (s *DailyTargetService) RefreshActualValues(userID uuid.UUID, targetID uint) (*models.DailyTarget, error) {
	target, err := s.GetDailyTargetByID(userID, targetID)
	if err != nil {
		return nil, err
	}

	if err := s.updateActualValues(target); err != nil {
		return nil, err
	}

	if err := s.db.Save(target).Error; err != nil {
		return nil, err
	}

	return target, nil
}

// GetCurrentMonthSummary retrieves summary for current month
func (s *DailyTargetService) GetCurrentMonthSummary(userID uuid.UUID) (map[string]interface{}, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	var targets []models.DailyTarget
	if err := s.db.Where("user_id = ? AND date >= ? AND date <= ?", userID, startOfMonth, endOfMonth).Find(&targets).Error; err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"total_days":           len(targets),
		"total_income_target":  0.0,
		"total_expense_limit":  0.0,
		"total_savings_target": 0.0,
		"total_actual_income":  0.0,
		"total_actual_expense": 0.0,
		"total_actual_savings": 0.0,
		"days_met_income":      0,
		"days_met_expense":     0,
		"days_met_savings":     0,
	}

	for _, target := range targets {
		summary["total_income_target"] = summary["total_income_target"].(float64) + target.IncomeTarget
		summary["total_expense_limit"] = summary["total_expense_limit"].(float64) + target.ExpenseLimit
		summary["total_savings_target"] = summary["total_savings_target"].(float64) + target.SavingsTarget
		summary["total_actual_income"] = summary["total_actual_income"].(float64) + target.ActualIncome
		summary["total_actual_expense"] = summary["total_actual_expense"].(float64) + target.ActualExpense
		summary["total_actual_savings"] = summary["total_actual_savings"].(float64) + target.ActualSavings

		if target.ActualIncome >= target.IncomeTarget && target.IncomeTarget > 0 {
			summary["days_met_income"] = summary["days_met_income"].(int) + 1
		}
		if target.ActualExpense <= target.ExpenseLimit && target.ExpenseLimit > 0 {
			summary["days_met_expense"] = summary["days_met_expense"].(int) + 1
		}
		if target.ActualSavings >= target.SavingsTarget && target.SavingsTarget > 0 {
			summary["days_met_savings"] = summary["days_met_savings"].(int) + 1
		}
	}

	return summary, nil
}

// updateActualValues calculates actual values from transactions for a specific date
func (s *DailyTargetService) updateActualValues(target *models.DailyTarget) error {
	startOfDay := time.Date(target.Date.Year(), target.Date.Month(), target.Date.Day(), 0, 0, 0, 0, target.Date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Second)

	summary, err := s.transactionService.GetTransactionSummary(target.UserID, &startOfDay, &endOfDay)
	if err != nil {
		return err
	}

	target.ActualIncome = summary.TotalIncome
	target.ActualExpense = summary.TotalExpense
	target.ActualSavings = summary.NetBalance

	return nil
}

// GetTodayTarget retrieves or creates today's target
func (s *DailyTargetService) GetTodayTarget(userID uuid.UUID) (*models.DailyTarget, error) {
	today := time.Now().Truncate(24 * time.Hour)

	target, err := s.GetDailyTargetByDate(userID, today)
	if err != nil {
		// If not found, create default target for today
		if err.Error() == "daily target not found" {
			return s.CreateDailyTarget(userID, &models.CreateDailyTargetRequest{
				Date:          today,
				IncomeTarget:  0,
				ExpenseLimit:  0,
				SavingsTarget: 0,
				Notes:         "Auto-created target",
			})
		}
		return nil, err
	}

	// Refresh actual values
	if err := s.updateActualValues(target); err != nil {
		return nil, err
	}

	if err := s.db.Save(target).Error; err != nil {
		return nil, err
	}

	return target, nil
}

// GetWeekSummary retrieves summary for current week
func (s *DailyTargetService) GetWeekSummary(userID uuid.UUID) (map[string]interface{}, error) {
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	var targets []models.DailyTarget
	if err := s.db.Where("user_id = ? AND date >= ? AND date <= ?", userID, startOfWeek, endOfWeek).Find(&targets).Error; err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"week_start":        startOfWeek.Format("2006-01-02"),
		"week_end":          endOfWeek.Format("2006-01-02"),
		"total_days":        len(targets),
		"total_income":      0.0,
		"total_expense":     0.0,
		"total_savings":     0.0,
		"avg_daily_income":  0.0,
		"avg_daily_expense": 0.0,
		"avg_daily_savings": 0.0,
	}

	totalIncome := 0.0
	totalExpense := 0.0
	totalSavings := 0.0

	for _, target := range targets {
		totalIncome += target.ActualIncome
		totalExpense += target.ActualExpense
		totalSavings += target.ActualSavings
	}

	summary["total_income"] = totalIncome
	summary["total_expense"] = totalExpense
	summary["total_savings"] = totalSavings

	if len(targets) > 0 {
		summary["avg_daily_income"] = totalIncome / float64(len(targets))
		summary["avg_daily_expense"] = totalExpense / float64(len(targets))
		summary["avg_daily_savings"] = totalSavings / float64(len(targets))
	}

	return summary, nil
}
