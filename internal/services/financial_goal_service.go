package services

import (
	"errors"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FinancialGoalService struct {
	db *gorm.DB
}

func NewFinancialGoalService(db *gorm.DB) *FinancialGoalService {
	return &FinancialGoalService{db: db}
}

// CreateFinancialGoal creates a new financial goal
func (s *FinancialGoalService) CreateFinancialGoal(userID uuid.UUID, req *models.CreateFinancialGoalRequest) (*models.FinancialGoal, error) {
	// Validate dates
	if req.TargetDate.Before(req.StartDate) {
		return nil, errors.New("target date must be after start date")
	}

	goal := &models.FinancialGoal{
		UserID:        userID,
		Title:         req.Title,
		Description:   req.Description,
		TargetAmount:  req.TargetAmount,
		CurrentAmount: req.CurrentAmount,
		StartDate:     req.StartDate,
		TargetDate:    req.TargetDate,
		Category:      req.Category,
		Priority:      req.Priority,
		IsCompleted:   false,
	}

	// Check if already completed
	if goal.CurrentAmount >= goal.TargetAmount {
		goal.IsCompleted = true
		now := time.Now()
		goal.CompletedAt = &now
	}

	if err := s.db.Create(goal).Error; err != nil {
		return nil, err
	}

	return goal, nil
}

// GetFinancialGoalByID retrieves a financial goal by ID
func (s *FinancialGoalService) GetFinancialGoalByID(userID uuid.UUID, goalID uint) (*models.FinancialGoal, error) {
	var goal models.FinancialGoal
	if err := s.db.Where("id = ? AND user_id = ?", goalID, userID).First(&goal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("financial goal not found")
		}
		return nil, err
	}
	return &goal, nil
}

// GetAllFinancialGoals retrieves all financial goals with filters
func (s *FinancialGoalService) GetAllFinancialGoals(userID uuid.UUID, page, pageSize int, category string, priority *int, isCompleted *bool) ([]models.FinancialGoal, int64, error) {
	var goals []models.FinancialGoal
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	query := s.db.Model(&models.FinancialGoal{}).Where("user_id = ?", userID)

	// Apply filters
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if priority != nil {
		query = query.Where("priority = ?", *priority)
	}
	if isCompleted != nil {
		query = query.Where("is_completed = ?", *isCompleted)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	offset := (page - 1) * pageSize
	if err := query.Order("priority DESC, target_date ASC").Limit(pageSize).Offset(offset).Find(&goals).Error; err != nil {
		return nil, 0, err
	}

	return goals, total, nil
}

// UpdateFinancialGoal updates a financial goal
func (s *FinancialGoalService) UpdateFinancialGoal(userID uuid.UUID, goalID uint, req *models.UpdateFinancialGoalRequest) (*models.FinancialGoal, error) {
	goal, err := s.GetFinancialGoalByID(userID, goalID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != "" {
		goal.Title = req.Title
	}
	if req.Description != "" {
		goal.Description = req.Description
	}
	if req.TargetAmount != nil {
		goal.TargetAmount = *req.TargetAmount
	}
	if req.CurrentAmount != nil {
		goal.CurrentAmount = *req.CurrentAmount
	}
	if req.StartDate != nil {
		goal.StartDate = *req.StartDate
	}
	if req.TargetDate != nil {
		goal.TargetDate = *req.TargetDate
	}
	if req.Category != "" {
		goal.Category = req.Category
	}
	if req.Priority != nil {
		goal.Priority = *req.Priority
	}
	if req.IsCompleted != nil {
		goal.IsCompleted = *req.IsCompleted
		if *req.IsCompleted && goal.CompletedAt == nil {
			now := time.Now()
			goal.CompletedAt = &now
		} else if !*req.IsCompleted {
			goal.CompletedAt = nil
		}
	}

	// Validate dates
	if goal.TargetDate.Before(goal.StartDate) {
		return nil, errors.New("target date must be after start date")
	}

	// Auto-complete if target reached
	if goal.CurrentAmount >= goal.TargetAmount && !goal.IsCompleted {
		goal.IsCompleted = true
		now := time.Now()
		goal.CompletedAt = &now
	}

	if err := s.db.Save(goal).Error; err != nil {
		return nil, err
	}

	return goal, nil
}

// DeleteFinancialGoal soft deletes a financial goal
func (s *FinancialGoalService) DeleteFinancialGoal(userID uuid.UUID, goalID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", goalID, userID).Delete(&models.FinancialGoal{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("financial goal not found")
	}
	return nil
}

// UpdateGoalProgress updates the current amount of a goal
func (s *FinancialGoalService) UpdateGoalProgress(userID uuid.UUID, goalID uint, amount float64) (*models.FinancialGoal, error) {
	goal, err := s.GetFinancialGoalByID(userID, goalID)
	if err != nil {
		return nil, err
	}

	if amount < 0 {
		return nil, errors.New("amount cannot be negative")
	}

	goal.CurrentAmount = amount

	// Auto-complete if target reached
	if goal.CurrentAmount >= goal.TargetAmount && !goal.IsCompleted {
		goal.IsCompleted = true
		now := time.Now()
		goal.CompletedAt = &now
	}

	if err := s.db.Save(goal).Error; err != nil {
		return nil, err
	}

	return goal, nil
}

// AddToGoalProgress adds amount to current goal progress
func (s *FinancialGoalService) AddToGoalProgress(userID uuid.UUID, goalID uint, amount float64) (*models.FinancialGoal, error) {
	goal, err := s.GetFinancialGoalByID(userID, goalID)
	if err != nil {
		return nil, err
	}

	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	goal.CurrentAmount += amount

	// Auto-complete if target reached
	if goal.CurrentAmount >= goal.TargetAmount && !goal.IsCompleted {
		goal.IsCompleted = true
		now := time.Now()
		goal.CompletedAt = &now
	}

	if err := s.db.Save(goal).Error; err != nil {
		return nil, err
	}

	return goal, nil
}

// GetActiveGoals retrieves all active (incomplete) goals
func (s *FinancialGoalService) GetActiveGoals(userID uuid.UUID, page, pageSize int) ([]models.FinancialGoal, int64, error) {
	isCompleted := false
	return s.GetAllFinancialGoals(userID, page, pageSize, "", nil, &isCompleted)
}

// GetCompletedGoals retrieves all completed goals
func (s *FinancialGoalService) GetCompletedGoals(userID uuid.UUID, page, pageSize int) ([]models.FinancialGoal, int64, error) {
	isCompleted := true
	return s.GetAllFinancialGoals(userID, page, pageSize, "", nil, &isCompleted)
}

// GetGoalsByPriority retrieves goals filtered by priority
func (s *FinancialGoalService) GetGoalsByPriority(userID uuid.UUID, priority int, page, pageSize int) ([]models.FinancialGoal, int64, error) {
	return s.GetAllFinancialGoals(userID, page, pageSize, "", &priority, nil)
}

// GetOverdueGoals retrieves goals that are past their target date and not completed
func (s *FinancialGoalService) GetOverdueGoals(userID uuid.UUID, page, pageSize int) ([]models.FinancialGoal, int64, error) {
	var goals []models.FinancialGoal
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	now := time.Now()
	query := s.db.Model(&models.FinancialGoal{}).
		Where("user_id = ? AND is_completed = ? AND target_date < ?", userID, false, now)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Order("target_date ASC").Limit(pageSize).Offset(offset).Find(&goals).Error; err != nil {
		return nil, 0, err
	}

	return goals, total, nil
}

// GetGoalsSummary retrieves summary statistics for all goals
func (s *FinancialGoalService) GetGoalsSummary(userID uuid.UUID) (map[string]interface{}, error) {
	var goals []models.FinancialGoal
	if err := s.db.Where("user_id = ?", userID).Find(&goals).Error; err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"total_goals":      len(goals),
		"active_goals":     0,
		"completed_goals":  0,
		"overdue_goals":    0,
		"total_target":     0.0,
		"total_saved":      0.0,
		"total_remaining":  0.0,
		"overall_progress": 0.0,
		"high_priority":    0,
		"medium_priority":  0,
		"low_priority":     0,
	}

	now := time.Now()
	totalTarget := 0.0
	totalSaved := 0.0

	for _, goal := range goals {
		if goal.IsCompleted {
			summary["completed_goals"] = summary["completed_goals"].(int) + 1
		} else {
			summary["active_goals"] = summary["active_goals"].(int) + 1
			if goal.TargetDate.Before(now) {
				summary["overdue_goals"] = summary["overdue_goals"].(int) + 1
			}
		}

		totalTarget += goal.TargetAmount
		totalSaved += goal.CurrentAmount

		switch goal.Priority {
		case 0:
			summary["low_priority"] = summary["low_priority"].(int) + 1
		case 1:
			summary["medium_priority"] = summary["medium_priority"].(int) + 1
		case 2:
			summary["high_priority"] = summary["high_priority"].(int) + 1
		}
	}

	summary["total_target"] = totalTarget
	summary["total_saved"] = totalSaved
	summary["total_remaining"] = totalTarget - totalSaved

	if totalTarget > 0 {
		summary["overall_progress"] = (totalSaved / totalTarget) * 100
	}

	return summary, nil
}
