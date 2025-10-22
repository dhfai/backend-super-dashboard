package services

import (
	"errors"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BudgetService struct {
	db                 *gorm.DB
	transactionService *TransactionService
}

func NewBudgetService(db *gorm.DB, transactionService *TransactionService) *BudgetService {
	return &BudgetService{
		db:                 db,
		transactionService: transactionService,
	}
}

// CreateBudget creates a new budget
func (s *BudgetService) CreateBudget(userID uuid.UUID, req *models.CreateBudgetRequest) (*models.Budget, error) {
	// Validate dates
	if !req.EndDate.IsZero() && req.EndDate.Before(req.StartDate) {
		return nil, errors.New("end date must be after start date")
	}

	budget := &models.Budget{
		UserID:      userID,
		Category:    req.Category,
		Amount:      req.Amount,
		Period:      req.Period,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		IsRecurring: req.IsRecurring,
	}

	// Calculate spent amount
	if err := s.updateSpentAmount(budget); err != nil {
		return nil, err
	}

	if err := s.db.Create(budget).Error; err != nil {
		return nil, err
	}

	return budget, nil
}

// GetBudgetByID retrieves a budget by ID
func (s *BudgetService) GetBudgetByID(userID uuid.UUID, budgetID uint) (*models.Budget, error) {
	var budget models.Budget
	if err := s.db.Where("id = ? AND user_id = ?", budgetID, userID).First(&budget).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("budget not found")
		}
		return nil, err
	}
	return &budget, nil
}

// GetAllBudgets retrieves all budgets with filters
func (s *BudgetService) GetAllBudgets(userID uuid.UUID, page, pageSize int, category, period string, isActive *bool) ([]models.Budget, int64, error) {
	var budgets []models.Budget
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	query := s.db.Model(&models.Budget{}).Where("user_id = ?", userID)

	// Apply filters
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if period != "" {
		query = query.Where("period = ?", period)
	}
	if isActive != nil && *isActive {
		now := time.Now()
		query = query.Where("(end_date IS NULL OR end_date >= ?) AND start_date <= ?", now, now)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Order("start_date DESC").Limit(pageSize).Offset(offset).Find(&budgets).Error; err != nil {
		return nil, 0, err
	}

	return budgets, total, nil
}

// UpdateBudget updates a budget
func (s *BudgetService) UpdateBudget(userID uuid.UUID, budgetID uint, req *models.UpdateBudgetRequest) (*models.Budget, error) {
	budget, err := s.GetBudgetByID(userID, budgetID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Category != "" {
		budget.Category = req.Category
	}
	if req.Amount != nil {
		budget.Amount = *req.Amount
	}
	if req.Period != "" {
		budget.Period = req.Period
	}
	if req.StartDate != nil {
		budget.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		budget.EndDate = *req.EndDate
	}
	if req.IsRecurring != nil {
		budget.IsRecurring = *req.IsRecurring
	}

	// Validate dates
	if !budget.EndDate.IsZero() && budget.EndDate.Before(budget.StartDate) {
		return nil, errors.New("end date must be after start date")
	}

	// Recalculate spent amount
	if err := s.updateSpentAmount(budget); err != nil {
		return nil, err
	}

	if err := s.db.Save(budget).Error; err != nil {
		return nil, err
	}

	return budget, nil
}

// DeleteBudget soft deletes a budget
func (s *BudgetService) DeleteBudget(userID uuid.UUID, budgetID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", budgetID, userID).Delete(&models.Budget{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("budget not found")
	}
	return nil
}

// RefreshSpentAmount updates spent amount from transactions
func (s *BudgetService) RefreshSpentAmount(userID uuid.UUID, budgetID uint) (*models.Budget, error) {
	budget, err := s.GetBudgetByID(userID, budgetID)
	if err != nil {
		return nil, err
	}

	if err := s.updateSpentAmount(budget); err != nil {
		return nil, err
	}

	if err := s.db.Save(budget).Error; err != nil {
		return nil, err
	}

	return budget, nil
}

// GetActiveBudgets retrieves all currently active budgets
func (s *BudgetService) GetActiveBudgets(userID uuid.UUID, page, pageSize int) ([]models.Budget, int64, error) {
	isActive := true
	return s.GetAllBudgets(userID, page, pageSize, "", "", &isActive)
}

// GetBudgetsByCategory retrieves budgets for a specific category
func (s *BudgetService) GetBudgetsByCategory(userID uuid.UUID, category string, page, pageSize int) ([]models.Budget, int64, error) {
	return s.GetAllBudgets(userID, page, pageSize, category, "", nil)
}

// GetOverBudgetCategories retrieves categories that have exceeded their budget
func (s *BudgetService) GetOverBudgetCategories(userID uuid.UUID) ([]models.Budget, error) {
	var budgets []models.Budget
	now := time.Now()

	if err := s.db.Where("user_id = ? AND (end_date IS NULL OR end_date >= ?) AND start_date <= ?", userID, now, now).Find(&budgets).Error; err != nil {
		return nil, err
	}

	var overBudget []models.Budget
	for _, budget := range budgets {
		if budget.SpentAmount > budget.Amount {
			overBudget = append(overBudget, budget)
		}
	}

	return overBudget, nil
}

// GetBudgetSummary retrieves summary of all budgets
func (s *BudgetService) GetBudgetSummary(userID uuid.UUID) (map[string]interface{}, error) {
	var budgets []models.Budget
	now := time.Now()

	if err := s.db.Where("user_id = ? AND (end_date IS NULL OR end_date >= ?) AND start_date <= ?", userID, now, now).Find(&budgets).Error; err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"total_budgets":      len(budgets),
		"total_allocated":    0.0,
		"total_spent":        0.0,
		"total_remaining":    0.0,
		"over_budget_count":  0,
		"under_budget_count": 0,
		"on_track_count":     0,
		"categories":         make(map[string]interface{}),
	}

	totalAllocated := 0.0
	totalSpent := 0.0
	categories := make(map[string]interface{})

	for _, budget := range budgets {
		totalAllocated += budget.Amount
		totalSpent += budget.SpentAmount

		if budget.SpentAmount > budget.Amount {
			summary["over_budget_count"] = summary["over_budget_count"].(int) + 1
		} else if budget.SpentAmount < budget.Amount*0.8 {
			summary["under_budget_count"] = summary["under_budget_count"].(int) + 1
		} else {
			summary["on_track_count"] = summary["on_track_count"].(int) + 1
		}

		// Category breakdown
		if _, exists := categories[budget.Category]; !exists {
			categories[budget.Category] = map[string]interface{}{
				"allocated": 0.0,
				"spent":     0.0,
				"remaining": 0.0,
			}
		}

		catData := categories[budget.Category].(map[string]interface{})
		catData["allocated"] = catData["allocated"].(float64) + budget.Amount
		catData["spent"] = catData["spent"].(float64) + budget.SpentAmount
		catData["remaining"] = catData["allocated"].(float64) - catData["spent"].(float64)
		categories[budget.Category] = catData
	}

	summary["total_allocated"] = totalAllocated
	summary["total_spent"] = totalSpent
	summary["total_remaining"] = totalAllocated - totalSpent
	summary["categories"] = categories

	return summary, nil
}

// CheckBudgetAlert checks if any budget is close to limit
func (s *BudgetService) CheckBudgetAlert(userID uuid.UUID, threshold float64) ([]map[string]interface{}, error) {
	var budgets []models.Budget
	now := time.Now()

	if err := s.db.Where("user_id = ? AND (end_date IS NULL OR end_date >= ?) AND start_date <= ?", userID, now, now).Find(&budgets).Error; err != nil {
		return nil, err
	}

	var alerts []map[string]interface{}
	for _, budget := range budgets {
		percentage := (budget.SpentAmount / budget.Amount) * 100
		if percentage >= threshold {
			alerts = append(alerts, map[string]interface{}{
				"budget_id":  budget.ID,
				"category":   budget.Category,
				"amount":     budget.Amount,
				"spent":      budget.SpentAmount,
				"remaining":  budget.Amount - budget.SpentAmount,
				"percentage": percentage,
				"status":     s.getBudgetStatus(percentage),
			})
		}
	}

	return alerts, nil
}

// updateSpentAmount calculates spent amount from transactions
func (s *BudgetService) updateSpentAmount(budget *models.Budget) error {
	endDate := budget.EndDate
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// Get transactions for this category in the date range
	filter := &models.TransactionFilterRequest{
		Type:     string(models.TransactionTypeExpense),
		Category: budget.Category,
		DateFrom: &budget.StartDate,
		DateTo:   &endDate,
		PageSize: 1000, // Get all transactions
	}

	transactions, _, err := s.transactionService.GetAllTransactions(budget.UserID, filter)
	if err != nil {
		return err
	}

	totalSpent := 0.0
	for _, transaction := range transactions {
		totalSpent += transaction.Amount
	}

	budget.SpentAmount = totalSpent
	return nil
}

// getBudgetStatus returns the status based on percentage
func (s *BudgetService) getBudgetStatus(percentage float64) string {
	if percentage >= 100 {
		return "exceeded"
	} else if percentage >= 90 {
		return "critical"
	} else if percentage >= 75 {
		return "warning"
	}
	return "good"
}

// GetMonthlyBudgetReport retrieves monthly budget report
func (s *BudgetService) GetMonthlyBudgetReport(userID uuid.UUID, year int, month int) (map[string]interface{}, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	var budgets []models.Budget
	if err := s.db.Where("user_id = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)", userID, endDate, startDate).Find(&budgets).Error; err != nil {
		return nil, err
	}

	report := map[string]interface{}{
		"year":            year,
		"month":           month,
		"total_budgets":   len(budgets),
		"total_allocated": 0.0,
		"total_spent":     0.0,
		"categories":      make([]map[string]interface{}, 0),
	}

	totalAllocated := 0.0
	totalSpent := 0.0

	for _, budget := range budgets {
		totalAllocated += budget.Amount
		totalSpent += budget.SpentAmount

		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (budget.SpentAmount / budget.Amount) * 100
		}

		report["categories"] = append(report["categories"].([]map[string]interface{}), map[string]interface{}{
			"category":   budget.Category,
			"allocated":  budget.Amount,
			"spent":      budget.SpentAmount,
			"remaining":  budget.Amount - budget.SpentAmount,
			"percentage": percentage,
			"status":     s.getBudgetStatus(percentage),
		})
	}

	report["total_allocated"] = totalAllocated
	report["total_spent"] = totalSpent
	report["total_remaining"] = totalAllocated - totalSpent

	return report, nil
}
