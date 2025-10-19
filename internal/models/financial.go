package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionType represents the type of financial transaction
type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

// Transaction represents a financial transaction (income or expense)
type Transaction struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	UserID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"user_id"`
	Type        TransactionType `gorm:"type:varchar(20);not null" json:"type"`
	Amount      float64         `gorm:"type:decimal(15,2);not null" json:"amount"`
	Category    string          `gorm:"type:varchar(100);not null" json:"category"`
	Description string          `gorm:"type:text" json:"description"`
	Date        time.Time       `gorm:"type:date;not null;index" json:"date"`
	Tags        string          `gorm:"type:varchar(500)" json:"tags"` // Comma-separated tags
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
	User        User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for Transaction model
func (Transaction) TableName() string {
	return "transactions"
}

// DailyTarget represents daily financial targets
type DailyTarget struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Date          time.Time      `gorm:"type:date;not null;index;uniqueIndex:idx_user_date" json:"date"`
	IncomeTarget  float64        `gorm:"type:decimal(15,2);not null;default:0" json:"income_target"`
	ExpenseLimit  float64        `gorm:"type:decimal(15,2);not null;default:0" json:"expense_limit"`
	SavingsTarget float64        `gorm:"type:decimal(15,2);not null;default:0" json:"savings_target"`
	ActualIncome  float64        `gorm:"type:decimal(15,2);default:0" json:"actual_income"`
	ActualExpense float64        `gorm:"type:decimal(15,2);default:0" json:"actual_expense"`
	ActualSavings float64        `gorm:"type:decimal(15,2);default:0" json:"actual_savings"`
	Notes         string         `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for DailyTarget model
func (DailyTarget) TableName() string {
	return "daily_targets"
}

// FinancialGoal represents long-term financial goals
type FinancialGoal struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`
	Description   string         `gorm:"type:text" json:"description"`
	TargetAmount  float64        `gorm:"type:decimal(15,2);not null" json:"target_amount"`
	CurrentAmount float64        `gorm:"type:decimal(15,2);default:0" json:"current_amount"`
	StartDate     time.Time      `gorm:"type:date;not null" json:"start_date"`
	TargetDate    time.Time      `gorm:"type:date;not null" json:"target_date"`
	Category      string         `gorm:"type:varchar(100)" json:"category"`
	Priority      int            `gorm:"default:0" json:"priority"` // 0=low, 1=medium, 2=high
	IsCompleted   bool           `gorm:"default:false" json:"is_completed"`
	CompletedAt   *time.Time     `gorm:"type:timestamp" json:"completed_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for FinancialGoal model
func (FinancialGoal) TableName() string {
	return "financial_goals"
}

// BacktestStrategy represents a financial strategy for backtesting
type BacktestStrategy struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	UserID              uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Name                string         `gorm:"type:varchar(255);not null" json:"name"`
	Description         string         `gorm:"type:text" json:"description"`
	StrategyType        string         `gorm:"type:varchar(100);not null" json:"strategy_type"` // savings, investment, expense_reduction
	InitialAmount       float64        `gorm:"type:decimal(15,2);not null" json:"initial_amount"`
	MonthlyContribution float64        `gorm:"type:decimal(15,2);default:0" json:"monthly_contribution"`
	ExpectedReturn      float64        `gorm:"type:decimal(5,2);default:0" json:"expected_return"` // percentage
	Duration            int            `gorm:"not null" json:"duration"`                           // in months
	StartDate           time.Time      `gorm:"type:date;not null" json:"start_date"`
	EndDate             time.Time      `gorm:"type:date;not null" json:"end_date"`
	ProjectedAmount     float64        `gorm:"type:decimal(15,2)" json:"projected_amount"`
	ActualAmount        float64        `gorm:"type:decimal(15,2);default:0" json:"actual_amount"`
	Status              string         `gorm:"type:varchar(50);default:'draft'" json:"status"` // draft, active, completed, cancelled
	ResultData          string         `gorm:"type:json" json:"result_data,omitempty"`         // JSON data for backtest results
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	User                User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for BacktestStrategy model
func (BacktestStrategy) TableName() string {
	return "backtest_strategies"
}

// Budget represents monthly or category-based budgets
type Budget struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Category    string         `gorm:"type:varchar(100);not null" json:"category"`
	Amount      float64        `gorm:"type:decimal(15,2);not null" json:"amount"`
	Period      string         `gorm:"type:varchar(50);not null" json:"period"` // monthly, weekly, yearly
	StartDate   time.Time      `gorm:"type:date;not null" json:"start_date"`
	EndDate     time.Time      `gorm:"type:date" json:"end_date,omitempty"`
	SpentAmount float64        `gorm:"type:decimal(15,2);default:0" json:"spent_amount"`
	IsRecurring bool           `gorm:"default:false" json:"is_recurring"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for Budget model
func (Budget) TableName() string {
	return "budgets"
}
