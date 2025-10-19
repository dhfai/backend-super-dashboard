package services

import (
	"errors"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionService struct {
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

// CreateTransaction creates a new transaction
func (s *TransactionService) CreateTransaction(userID uuid.UUID, req *models.CreateTransactionRequest) (*models.Transaction, error) {
	transaction := &models.Transaction{
		UserID:      userID,
		Type:        models.TransactionType(req.Type),
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
		Tags:        req.Tags,
	}

	if err := s.db.Create(transaction).Error; err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetTransactionByID retrieves a transaction by ID
func (s *TransactionService) GetTransactionByID(userID uuid.UUID, transactionID uint) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := s.db.Where("id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

// GetAllTransactions retrieves all transactions with filters
func (s *TransactionService) GetAllTransactions(userID uuid.UUID, filter *models.TransactionFilterRequest) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	// Set default values
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.SortBy == "" {
		filter.SortBy = "date"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	query := s.db.Model(&models.Transaction{}).Where("user_id = ?", userID)

	// Apply filters
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Tags != "" {
		query = query.Where("tags LIKE ?", "%"+filter.Tags+"%")
	}
	if filter.DateFrom != nil {
		query = query.Where("date >= ?", filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("date <= ?", filter.DateTo)
	}
	if filter.MinAmount != nil {
		query = query.Where("amount >= ?", *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		query = query.Where("amount <= ?", *filter.MaxAmount)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting and pagination
	offset := (filter.Page - 1) * filter.PageSize
	orderClause := filter.SortBy + " " + filter.SortOrder
	if err := query.Order(orderClause).Limit(filter.PageSize).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// UpdateTransaction updates a transaction
func (s *TransactionService) UpdateTransaction(userID uuid.UUID, transactionID uint, req *models.UpdateTransactionRequest) (*models.Transaction, error) {
	transaction, err := s.GetTransactionByID(userID, transactionID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Type != "" {
		transaction.Type = models.TransactionType(req.Type)
	}
	if req.Amount != nil {
		transaction.Amount = *req.Amount
	}
	if req.Category != "" {
		transaction.Category = req.Category
	}
	if req.Description != "" {
		transaction.Description = req.Description
	}
	if req.Date != nil {
		transaction.Date = *req.Date
	}
	if req.Tags != "" {
		transaction.Tags = req.Tags
	}

	if err := s.db.Save(transaction).Error; err != nil {
		return nil, err
	}

	return transaction, nil
}

// DeleteTransaction soft deletes a transaction
func (s *TransactionService) DeleteTransaction(userID uuid.UUID, transactionID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", transactionID, userID).Delete(&models.Transaction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("transaction not found")
	}
	return nil
}

// HardDeleteTransaction permanently deletes a transaction
func (s *TransactionService) HardDeleteTransaction(userID uuid.UUID, transactionID uint) error {
	result := s.db.Unscoped().Where("id = ? AND user_id = ?", transactionID, userID).Delete(&models.Transaction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("transaction not found")
	}
	return nil
}

// GetTransactionsByCategory retrieves transactions by category
func (s *TransactionService) GetTransactionsByCategory(userID uuid.UUID, category string, page, pageSize int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	query := s.db.Model(&models.Transaction{}).Where("user_id = ? AND category = ?", userID, category)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("date DESC").Limit(pageSize).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetTransactionSummary calculates summary of transactions
func (s *TransactionService) GetTransactionSummary(userID uuid.UUID, dateFrom, dateTo *time.Time) (*models.TransactionSummary, error) {
	summary := &models.TransactionSummary{}

	query := s.db.Model(&models.Transaction{}).Where("user_id = ?", userID)

	if dateFrom != nil {
		query = query.Where("date >= ?", dateFrom)
	}
	if dateTo != nil {
		query = query.Where("date <= ?", dateTo)
	}

	// Calculate total income
	var totalIncome float64
	if err := query.Where("type = ?", models.TransactionTypeIncome).Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome).Error; err != nil {
		return nil, err
	}
	summary.TotalIncome = totalIncome

	// Calculate total expense
	var totalExpense float64
	if err := query.Where("type = ?", models.TransactionTypeExpense).Select("COALESCE(SUM(amount), 0)").Scan(&totalExpense).Error; err != nil {
		return nil, err
	}
	summary.TotalExpense = totalExpense

	// Calculate net balance
	summary.NetBalance = totalIncome - totalExpense

	return summary, nil
}

// GetCategoryBreakdown retrieves spending breakdown by category
func (s *TransactionService) GetCategoryBreakdown(userID uuid.UUID, transactionType string, dateFrom, dateTo *time.Time) (map[string]float64, error) {
	type CategorySum struct {
		Category string
		Total    float64
	}

	var results []CategorySum
	query := s.db.Model(&models.Transaction{}).
		Select("category, SUM(amount) as total").
		Where("user_id = ?", userID).
		Group("category")

	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}
	if dateFrom != nil {
		query = query.Where("date >= ?", dateFrom)
	}
	if dateTo != nil {
		query = query.Where("date <= ?", dateTo)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	breakdown := make(map[string]float64)
	for _, result := range results {
		breakdown[result.Category] = result.Total
	}

	return breakdown, nil
}

// GetMonthlyTrend retrieves monthly transaction trends
func (s *TransactionService) GetMonthlyTrend(userID uuid.UUID, months int) ([]map[string]interface{}, error) {
	type MonthlyData struct {
		Month      string
		Income     float64
		Expense    float64
		NetBalance float64
	}

	var results []MonthlyData

	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	// PostgreSQL query for monthly aggregation
	query := `
		SELECT
			TO_CHAR(date, 'YYYY-MM') as month,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as expense,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0) as net_balance
		FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ? AND deleted_at IS NULL
		GROUP BY TO_CHAR(date, 'YYYY-MM')
		ORDER BY month ASC
	`

	if err := s.db.Raw(query, userID, startDate, endDate).Scan(&results).Error; err != nil {
		return nil, err
	}

	// Convert to interface map
	trend := make([]map[string]interface{}, len(results))
	for i, result := range results {
		trend[i] = map[string]interface{}{
			"month":       result.Month,
			"income":      result.Income,
			"expense":     result.Expense,
			"net_balance": result.NetBalance,
		}
	}

	return trend, nil
}

// GetTransactionsCount retrieves the total count of transactions
func (s *TransactionService) GetTransactionsCount(userID uuid.UUID) (int64, error) {
	var count int64
	if err := s.db.Model(&models.Transaction{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
