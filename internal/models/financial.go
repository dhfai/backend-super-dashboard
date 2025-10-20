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

// DailyTarget represents daily financial targets (for trading)
type DailyTarget struct {
	ID                uint              `gorm:"primaryKey" json:"id"`
	UserID            uuid.UUID         `gorm:"type:uuid;not null;index" json:"user_id"`
	Date              time.Time         `gorm:"type:date;not null;index;uniqueIndex:idx_user_date" json:"date"`
	IncomeTarget      float64           `gorm:"type:decimal(15,2);not null;default:0" json:"income_target"`  // Profit target
	ExpenseLimit      float64           `gorm:"type:decimal(15,2);not null;default:0" json:"expense_limit"`  // Max loss allowed
	SavingsTarget     float64           `gorm:"type:decimal(15,2);not null;default:0" json:"savings_target"` // Optional savings goal
	ActualIncome      float64           `gorm:"type:decimal(15,2);default:0" json:"actual_income"`           // Total profit achieved
	ActualExpense     float64           `gorm:"type:decimal(15,2);default:0" json:"actual_expense"`          // Total loss incurred
	ActualSavings     float64           `gorm:"type:decimal(15,2);default:0" json:"actual_savings"`          // Net profit (income - expense)
	RemainingIncome   float64           `gorm:"type:decimal(15,2);default:0" json:"remaining_income"`        // Income left to achieve
	RemainingExpense  float64           `gorm:"type:decimal(15,2);default:0" json:"remaining_expense"`       // Loss budget left
	TotalTrades       int               `gorm:"default:0" json:"total_trades"`                               // Number of trades
	WinningTrades     int               `gorm:"default:0" json:"winning_trades"`                             // Number of wins
	LosingTrades      int               `gorm:"default:0" json:"losing_trades"`                              // Number of losses
	WinRate           float64           `gorm:"type:decimal(5,2);default:0" json:"win_rate"`                 // Win rate percentage
	IsCompleted       bool              `gorm:"default:false" json:"is_completed"`                           // Target reached?
	CompletedAt       *time.Time        `gorm:"type:timestamp" json:"completed_at,omitempty"`                // When target was reached
	Notes             string            `gorm:"type:text" json:"notes"`                                      // Trading plan notes
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
	User              User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TradingActivities []TradingActivity `gorm:"foreignKey:DailyTargetID" json:"trading_activities,omitempty"`
}

// TableName specifies the table name for DailyTarget model
func (DailyTarget) TableName() string {
	return "daily_targets"
}

// TradingActivity represents individual trading activities linked to a daily target
type TradingActivity struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	DailyTargetID uint           `gorm:"not null;index" json:"daily_target_id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	TradeType     string         `gorm:"type:varchar(20);not null" json:"trade_type"`  // "win" or "loss"
	Amount        float64        `gorm:"type:decimal(15,2);not null" json:"amount"`    // Profit or loss amount
	Pips          int            `gorm:"default:0" json:"pips"`                        // Pips gained/lost
	LotSize       float64        `gorm:"type:decimal(10,2);default:0" json:"lot_size"` // Lot size (e.g., 0.01)
	Symbol        string         `gorm:"type:varchar(50)" json:"symbol"`               // Trading pair (e.g., EURUSD)
	Description   string         `gorm:"type:text" json:"description"`
	TradeTime     time.Time      `gorm:"type:timestamp;not null" json:"trade_time"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	DailyTarget   DailyTarget    `gorm:"foreignKey:DailyTargetID" json:"daily_target,omitempty"`
}

// TableName specifies the table name for TradingActivity model
func (TradingActivity) TableName() string {
	return "trading_activities"
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
