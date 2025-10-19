package services

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BacktestStrategyService struct {
	db *gorm.DB
}

func NewBacktestStrategyService(db *gorm.DB) *BacktestStrategyService {
	return &BacktestStrategyService{db: db}
}

// MonthlyResult represents monthly backtest calculation result
type MonthlyResult struct {
	Month        int     `json:"month"`
	StartBalance float64 `json:"start_balance"`
	Contribution float64 `json:"contribution"`
	Returns      float64 `json:"returns"`
	EndBalance   float64 `json:"end_balance"`
	TotalReturns float64 `json:"total_returns"`
}

// BacktestResult represents the complete backtest calculation
type BacktestResult struct {
	InitialAmount       float64         `json:"initial_amount"`
	MonthlyContribution float64         `json:"monthly_contribution"`
	ExpectedReturn      float64         `json:"expected_return"`
	Duration            int             `json:"duration"`
	TotalContributed    float64         `json:"total_contributed"`
	TotalReturns        float64         `json:"total_returns"`
	FinalAmount         float64         `json:"final_amount"`
	MonthlyResults      []MonthlyResult `json:"monthly_results"`
}

// CreateBacktestStrategy creates a new backtest strategy
func (s *BacktestStrategyService) CreateBacktestStrategy(userID uuid.UUID, req *models.CreateBacktestStrategyRequest) (*models.BacktestStrategy, error) {
	// Calculate end date
	endDate := req.StartDate.AddDate(0, req.Duration, 0)

	strategy := &models.BacktestStrategy{
		UserID:              userID,
		Name:                req.Name,
		Description:         req.Description,
		StrategyType:        req.StrategyType,
		InitialAmount:       req.InitialAmount,
		MonthlyContribution: req.MonthlyContribution,
		ExpectedReturn:      req.ExpectedReturn,
		Duration:            req.Duration,
		StartDate:           req.StartDate,
		EndDate:             endDate,
		Status:              "draft",
	}

	// Calculate projected amount
	result := s.calculateBacktest(strategy)
	strategy.ProjectedAmount = result.FinalAmount

	// Store result data as JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	strategy.ResultData = string(resultJSON)

	if err := s.db.Create(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

// GetBacktestStrategyByID retrieves a backtest strategy by ID
func (s *BacktestStrategyService) GetBacktestStrategyByID(userID uuid.UUID, strategyID uint) (*models.BacktestStrategy, error) {
	var strategy models.BacktestStrategy
	if err := s.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("backtest strategy not found")
		}
		return nil, err
	}
	return &strategy, nil
}

// GetAllBacktestStrategies retrieves all backtest strategies with filters
func (s *BacktestStrategyService) GetAllBacktestStrategies(userID uuid.UUID, page, pageSize int, strategyType, status string) ([]models.BacktestStrategy, int64, error) {
	var strategies []models.BacktestStrategy
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	query := s.db.Model(&models.BacktestStrategy{}).Where("user_id = ?", userID)

	// Apply filters
	if strategyType != "" {
		query = query.Where("strategy_type = ?", strategyType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&strategies).Error; err != nil {
		return nil, 0, err
	}

	return strategies, total, nil
}

// UpdateBacktestStrategy updates a backtest strategy
func (s *BacktestStrategyService) UpdateBacktestStrategy(userID uuid.UUID, strategyID uint, req *models.UpdateBacktestStrategyRequest) (*models.BacktestStrategy, error) {
	strategy, err := s.GetBacktestStrategyByID(userID, strategyID)
	if err != nil {
		return nil, err
	}

	needsRecalculation := false

	// Update fields if provided
	if req.Name != "" {
		strategy.Name = req.Name
	}
	if req.Description != "" {
		strategy.Description = req.Description
	}
	if req.StrategyType != "" {
		strategy.StrategyType = req.StrategyType
		needsRecalculation = true
	}
	if req.InitialAmount != nil {
		strategy.InitialAmount = *req.InitialAmount
		needsRecalculation = true
	}
	if req.MonthlyContribution != nil {
		strategy.MonthlyContribution = *req.MonthlyContribution
		needsRecalculation = true
	}
	if req.ExpectedReturn != nil {
		strategy.ExpectedReturn = *req.ExpectedReturn
		needsRecalculation = true
	}
	if req.Duration != nil {
		strategy.Duration = *req.Duration
		strategy.EndDate = strategy.StartDate.AddDate(0, strategy.Duration, 0)
		needsRecalculation = true
	}
	if req.Status != "" {
		strategy.Status = req.Status
	}
	if req.ActualAmount != nil {
		strategy.ActualAmount = *req.ActualAmount
	}

	// Recalculate if needed
	if needsRecalculation {
		result := s.calculateBacktest(strategy)
		strategy.ProjectedAmount = result.FinalAmount

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		strategy.ResultData = string(resultJSON)
	}

	if err := s.db.Save(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

// DeleteBacktestStrategy soft deletes a backtest strategy
func (s *BacktestStrategyService) DeleteBacktestStrategy(userID uuid.UUID, strategyID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", strategyID, userID).Delete(&models.BacktestStrategy{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("backtest strategy not found")
	}
	return nil
}

// RunBacktest recalculates the backtest for a strategy
func (s *BacktestStrategyService) RunBacktest(userID uuid.UUID, strategyID uint) (*models.BacktestStrategy, error) {
	strategy, err := s.GetBacktestStrategyByID(userID, strategyID)
	if err != nil {
		return nil, err
	}

	result := s.calculateBacktest(strategy)
	strategy.ProjectedAmount = result.FinalAmount

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	strategy.ResultData = string(resultJSON)

	if err := s.db.Save(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

// ActivateStrategy activates a draft strategy
func (s *BacktestStrategyService) ActivateStrategy(userID uuid.UUID, strategyID uint) (*models.BacktestStrategy, error) {
	strategy, err := s.GetBacktestStrategyByID(userID, strategyID)
	if err != nil {
		return nil, err
	}

	if strategy.Status != "draft" {
		return nil, errors.New("only draft strategies can be activated")
	}

	strategy.Status = "active"
	if err := s.db.Save(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

// CompleteStrategy marks a strategy as completed
func (s *BacktestStrategyService) CompleteStrategy(userID uuid.UUID, strategyID uint, actualAmount float64) (*models.BacktestStrategy, error) {
	strategy, err := s.GetBacktestStrategyByID(userID, strategyID)
	if err != nil {
		return nil, err
	}

	if strategy.Status != "active" {
		return nil, errors.New("only active strategies can be completed")
	}

	strategy.Status = "completed"
	strategy.ActualAmount = actualAmount

	if err := s.db.Save(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

// CancelStrategy cancels a strategy
func (s *BacktestStrategyService) CancelStrategy(userID uuid.UUID, strategyID uint) (*models.BacktestStrategy, error) {
	strategy, err := s.GetBacktestStrategyByID(userID, strategyID)
	if err != nil {
		return nil, err
	}

	if strategy.Status == "completed" {
		return nil, errors.New("completed strategies cannot be cancelled")
	}

	strategy.Status = "cancelled"
	if err := s.db.Save(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

// GetStrategiesByType retrieves strategies filtered by type
func (s *BacktestStrategyService) GetStrategiesByType(userID uuid.UUID, strategyType string, page, pageSize int) ([]models.BacktestStrategy, int64, error) {
	return s.GetAllBacktestStrategies(userID, page, pageSize, strategyType, "")
}

// GetActiveStrategies retrieves all active strategies
func (s *BacktestStrategyService) GetActiveStrategies(userID uuid.UUID, page, pageSize int) ([]models.BacktestStrategy, int64, error) {
	return s.GetAllBacktestStrategies(userID, page, pageSize, "", "active")
}

// GetCompletedStrategies retrieves all completed strategies
func (s *BacktestStrategyService) GetCompletedStrategies(userID uuid.UUID, page, pageSize int) ([]models.BacktestStrategy, int64, error) {
	return s.GetAllBacktestStrategies(userID, page, pageSize, "", "completed")
}

// CompareStrategies compares multiple strategies
func (s *BacktestStrategyService) CompareStrategies(userID uuid.UUID, strategyIDs []uint) ([]map[string]interface{}, error) {
	var strategies []models.BacktestStrategy
	if err := s.db.Where("user_id = ? AND id IN ?", userID, strategyIDs).Find(&strategies).Error; err != nil {
		return nil, err
	}

	comparison := make([]map[string]interface{}, len(strategies))
	for i, strategy := range strategies {
		comparison[i] = map[string]interface{}{
			"id":                   strategy.ID,
			"name":                 strategy.Name,
			"strategy_type":        strategy.StrategyType,
			"initial_amount":       strategy.InitialAmount,
			"monthly_contribution": strategy.MonthlyContribution,
			"expected_return":      strategy.ExpectedReturn,
			"duration":             strategy.Duration,
			"projected_amount":     strategy.ProjectedAmount,
			"actual_amount":        strategy.ActualAmount,
			"total_contribution":   strategy.InitialAmount + (strategy.MonthlyContribution * float64(strategy.Duration)),
			"total_gain":           strategy.ProjectedAmount - (strategy.InitialAmount + (strategy.MonthlyContribution * float64(strategy.Duration))),
			"roi_percentage":       ((strategy.ProjectedAmount - (strategy.InitialAmount + (strategy.MonthlyContribution * float64(strategy.Duration)))) / (strategy.InitialAmount + (strategy.MonthlyContribution * float64(strategy.Duration)))) * 100,
		}
	}

	return comparison, nil
}

// calculateBacktest performs the backtest calculation
func (s *BacktestStrategyService) calculateBacktest(strategy *models.BacktestStrategy) *BacktestResult {
	result := &BacktestResult{
		InitialAmount:       strategy.InitialAmount,
		MonthlyContribution: strategy.MonthlyContribution,
		ExpectedReturn:      strategy.ExpectedReturn,
		Duration:            strategy.Duration,
		MonthlyResults:      make([]MonthlyResult, strategy.Duration),
	}

	// Monthly return rate (annual return / 12)
	monthlyReturnRate := strategy.ExpectedReturn / 100 / 12

	balance := strategy.InitialAmount
	totalContributed := strategy.InitialAmount
	totalReturns := 0.0

	for month := 0; month < strategy.Duration; month++ {
		startBalance := balance

		// Add monthly contribution
		contribution := strategy.MonthlyContribution
		if month > 0 { // Don't add contribution in first month (initial amount already included)
			balance += contribution
			totalContributed += contribution
		}

		// Calculate returns
		returns := balance * monthlyReturnRate
		balance += returns
		totalReturns += returns

		result.MonthlyResults[month] = MonthlyResult{
			Month:        month + 1,
			StartBalance: math.Round(startBalance*100) / 100,
			Contribution: math.Round(contribution*100) / 100,
			Returns:      math.Round(returns*100) / 100,
			EndBalance:   math.Round(balance*100) / 100,
			TotalReturns: math.Round(totalReturns*100) / 100,
		}
	}

	result.TotalContributed = math.Round(totalContributed*100) / 100
	result.TotalReturns = math.Round(totalReturns*100) / 100
	result.FinalAmount = math.Round(balance*100) / 100

	return result
}

// GetBacktestResult retrieves the parsed backtest result
func (s *BacktestStrategyService) GetBacktestResult(userID uuid.UUID, strategyID uint) (*BacktestResult, error) {
	strategy, err := s.GetBacktestStrategyByID(userID, strategyID)
	if err != nil {
		return nil, err
	}

	if strategy.ResultData == "" {
		return nil, errors.New("no backtest result available")
	}

	var result BacktestResult
	if err := json.Unmarshal([]byte(strategy.ResultData), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
