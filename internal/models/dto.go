package models

import "time"

type RegisterRequest struct {
	Username       string `json:"username" validate:"required,min=3,max=50"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	RetypePassword string `json:"retype_password" validate:"required,eqfield=Password"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ForgetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTPCode     string `json:"otp_code" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type VerifyEmailRequest struct {
	Email   string `json:"email" validate:"required,email"`
	OTPCode string `json:"otp_code" validate:"required,len=6"`
}

type VerifyRegistrationRequest struct {
	Email   string `json:"email" validate:"required,email"`
	OTPCode string `json:"otp_code" validate:"required,len=6"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type DeleteAccountRequest struct {
	OTPCode string `json:"otp_code" validate:"required,len=6"`
}

type RequestDeleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	FullName    string `json:"full_name" validate:"max=100"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number" validate:"max=20"`
	Country     string `json:"country" validate:"max=50"`
}

type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

type UserResponse struct {
	ID              string     `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	IsActive        bool       `json:"is_active"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	Profile         *Profile   `json:"profile,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (u *User) ToUserResponse() *UserResponse {
	return &UserResponse{
		ID:              u.ID.String(),
		Username:        u.Username,
		Email:           u.Email,
		IsActive:        u.IsActive,
		EmailVerified:   u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		Profile:         u.Profile,
	}
}

// Note DTOs
type CreateNoteRequest struct {
	Title      string `json:"title" validate:"required,min=1,max=255"`
	Content    string `json:"content" validate:"required,min=1"`
	Tags       string `json:"tags" validate:"max=500"`
	IsFavorite bool   `json:"is_favorite"`
}

type UpdateNoteRequest struct {
	Title      string `json:"title" validate:"omitempty,min=1,max=255"`
	Content    string `json:"content" validate:"omitempty,min=1"`
	Tags       string `json:"tags" validate:"max=500"`
	IsFavorite *bool  `json:"is_favorite"`
}

type NoteResponse struct {
	ID         uint      `json:"id"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Tags       string    `json:"tags"`
	IsFavorite bool      `json:"is_favorite"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type NotesListResponse struct {
	Notes      []NoteResponse `json:"notes"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type NoteFilterRequest struct {
	Search     string `json:"search" form:"search"`
	Tags       string `json:"tags" form:"tags"`
	IsFavorite *bool  `json:"is_favorite" form:"is_favorite"`
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	SortBy     string `json:"sort_by" form:"sort_by"`       // created_at, updated_at, title
	SortOrder  string `json:"sort_order" form:"sort_order"` // asc, desc
}

func (n *Note) ToNoteResponse() *NoteResponse {
	return &NoteResponse{
		ID:         n.ID,
		UserID:     n.UserID.String(),
		Title:      n.Title,
		Content:    n.Content,
		Tags:       n.Tags,
		IsFavorite: n.IsFavorite,
		CreatedAt:  n.CreatedAt,
		UpdatedAt:  n.UpdatedAt,
	}
}

// Financial DTOs

// Transaction DTOs
type CreateTransactionRequest struct {
	Type        string    `json:"type" validate:"required,oneof=income expense"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Category    string    `json:"category" validate:"required,min=1,max=100"`
	Description string    `json:"description" validate:"max=1000"`
	Date        time.Time `json:"date" validate:"required"`
	Tags        string    `json:"tags" validate:"max=500"`
}

type UpdateTransactionRequest struct {
	Type        string     `json:"type" validate:"omitempty,oneof=income expense"`
	Amount      *float64   `json:"amount" validate:"omitempty,gt=0"`
	Category    string     `json:"category" validate:"omitempty,min=1,max=100"`
	Description string     `json:"description" validate:"max=1000"`
	Date        *time.Time `json:"date"`
	Tags        string     `json:"tags" validate:"max=500"`
}

type TransactionResponse struct {
	ID          uint      `json:"id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Tags        string    `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TransactionFilterRequest struct {
	Type      string     `json:"type" form:"type"` // income, expense
	Category  string     `json:"category" form:"category"`
	Tags      string     `json:"tags" form:"tags"`
	DateFrom  *time.Time `json:"date_from" form:"date_from"`
	DateTo    *time.Time `json:"date_to" form:"date_to"`
	MinAmount *float64   `json:"min_amount" form:"min_amount"`
	MaxAmount *float64   `json:"max_amount" form:"max_amount"`
	Page      int        `json:"page" form:"page"`
	PageSize  int        `json:"page_size" form:"page_size"`
	SortBy    string     `json:"sort_by" form:"sort_by"`       // date, amount, created_at
	SortOrder string     `json:"sort_order" form:"sort_order"` // asc, desc
}

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	PageSize     int                   `json:"page_size"`
	TotalPages   int                   `json:"total_pages"`
	Summary      *TransactionSummary   `json:"summary,omitempty"`
}

type TransactionSummary struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	NetBalance   float64 `json:"net_balance"`
}

func (t *Transaction) ToTransactionResponse() *TransactionResponse {
	return &TransactionResponse{
		ID:          t.ID,
		UserID:      t.UserID.String(),
		Type:        string(t.Type),
		Amount:      t.Amount,
		Category:    t.Category,
		Description: t.Description,
		Date:        t.Date,
		Tags:        t.Tags,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// DailyTarget DTOs
type CreateDailyTargetRequest struct {
	Date          time.Time `json:"date" validate:"required"`
	IncomeTarget  float64   `json:"income_target" validate:"gte=0"`
	ExpenseLimit  float64   `json:"expense_limit" validate:"gte=0"`
	SavingsTarget float64   `json:"savings_target" validate:"gte=0"`
	Notes         string    `json:"notes" validate:"max=1000"`
}

type UpdateDailyTargetRequest struct {
	IncomeTarget  *float64 `json:"income_target" validate:"omitempty,gte=0"`
	ExpenseLimit  *float64 `json:"expense_limit" validate:"omitempty,gte=0"`
	SavingsTarget *float64 `json:"savings_target" validate:"omitempty,gte=0"`
	Notes         string   `json:"notes" validate:"max=1000"`
}

type DailyTargetResponse struct {
	ID              uint      `json:"id"`
	UserID          string    `json:"user_id"`
	Date            time.Time `json:"date"`
	IncomeTarget    float64   `json:"income_target"`
	ExpenseLimit    float64   `json:"expense_limit"`
	SavingsTarget   float64   `json:"savings_target"`
	ActualIncome    float64   `json:"actual_income"`
	ActualExpense   float64   `json:"actual_expense"`
	ActualSavings   float64   `json:"actual_savings"`
	IncomeProgress  float64   `json:"income_progress"`  // percentage
	ExpenseProgress float64   `json:"expense_progress"` // percentage
	SavingsProgress float64   `json:"savings_progress"` // percentage
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type DailyTargetListResponse struct {
	Targets    []DailyTargetResponse `json:"targets"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

func (d *DailyTarget) ToDailyTargetResponse() *DailyTargetResponse {
	response := &DailyTargetResponse{
		ID:            d.ID,
		UserID:        d.UserID.String(),
		Date:          d.Date,
		IncomeTarget:  d.IncomeTarget,
		ExpenseLimit:  d.ExpenseLimit,
		SavingsTarget: d.SavingsTarget,
		ActualIncome:  d.ActualIncome,
		ActualExpense: d.ActualExpense,
		ActualSavings: d.ActualSavings,
		Notes:         d.Notes,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}

	// Calculate progress percentages
	if d.IncomeTarget > 0 {
		response.IncomeProgress = (d.ActualIncome / d.IncomeTarget) * 100
	}
	if d.ExpenseLimit > 0 {
		response.ExpenseProgress = (d.ActualExpense / d.ExpenseLimit) * 100
	}
	if d.SavingsTarget > 0 {
		response.SavingsProgress = (d.ActualSavings / d.SavingsTarget) * 100
	}

	return response
}

// FinancialGoal DTOs
type CreateFinancialGoalRequest struct {
	Title         string    `json:"title" validate:"required,min=1,max=255"`
	Description   string    `json:"description" validate:"max=1000"`
	TargetAmount  float64   `json:"target_amount" validate:"required,gt=0"`
	CurrentAmount float64   `json:"current_amount" validate:"gte=0"`
	StartDate     time.Time `json:"start_date" validate:"required"`
	TargetDate    time.Time `json:"target_date" validate:"required"`
	Category      string    `json:"category" validate:"max=100"`
	Priority      int       `json:"priority" validate:"gte=0,lte=2"`
}

type UpdateFinancialGoalRequest struct {
	Title         string     `json:"title" validate:"omitempty,min=1,max=255"`
	Description   string     `json:"description" validate:"max=1000"`
	TargetAmount  *float64   `json:"target_amount" validate:"omitempty,gt=0"`
	CurrentAmount *float64   `json:"current_amount" validate:"omitempty,gte=0"`
	StartDate     *time.Time `json:"start_date"`
	TargetDate    *time.Time `json:"target_date"`
	Category      string     `json:"category" validate:"max=100"`
	Priority      *int       `json:"priority" validate:"omitempty,gte=0,lte=2"`
	IsCompleted   *bool      `json:"is_completed"`
}

type FinancialGoalResponse struct {
	ID             uint       `json:"id"`
	UserID         string     `json:"user_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	TargetAmount   float64    `json:"target_amount"`
	CurrentAmount  float64    `json:"current_amount"`
	Progress       float64    `json:"progress"` // percentage
	StartDate      time.Time  `json:"start_date"`
	TargetDate     time.Time  `json:"target_date"`
	Category       string     `json:"category"`
	Priority       int        `json:"priority"`
	PriorityLabel  string     `json:"priority_label"`
	IsCompleted    bool       `json:"is_completed"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	DaysRemaining  int        `json:"days_remaining"`
	RequiredPerDay float64    `json:"required_per_day"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type FinancialGoalListResponse struct {
	Goals      []FinancialGoalResponse `json:"goals"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

func (f *FinancialGoal) ToFinancialGoalResponse() *FinancialGoalResponse {
	response := &FinancialGoalResponse{
		ID:            f.ID,
		UserID:        f.UserID.String(),
		Title:         f.Title,
		Description:   f.Description,
		TargetAmount:  f.TargetAmount,
		CurrentAmount: f.CurrentAmount,
		StartDate:     f.StartDate,
		TargetDate:    f.TargetDate,
		Category:      f.Category,
		Priority:      f.Priority,
		IsCompleted:   f.IsCompleted,
		CompletedAt:   f.CompletedAt,
		CreatedAt:     f.CreatedAt,
		UpdatedAt:     f.UpdatedAt,
	}

	// Calculate progress
	if f.TargetAmount > 0 {
		response.Progress = (f.CurrentAmount / f.TargetAmount) * 100
	}

	// Priority label
	switch f.Priority {
	case 0:
		response.PriorityLabel = "low"
	case 1:
		response.PriorityLabel = "medium"
	case 2:
		response.PriorityLabel = "high"
	}

	// Days remaining
	now := time.Now()
	if f.TargetDate.After(now) && !f.IsCompleted {
		response.DaysRemaining = int(f.TargetDate.Sub(now).Hours() / 24)

		// Required per day
		remaining := f.TargetAmount - f.CurrentAmount
		if remaining > 0 && response.DaysRemaining > 0 {
			response.RequiredPerDay = remaining / float64(response.DaysRemaining)
		}
	}

	return response
}

// BacktestStrategy DTOs
type CreateBacktestStrategyRequest struct {
	Name                string    `json:"name" validate:"required,min=1,max=255"`
	Description         string    `json:"description" validate:"max=1000"`
	StrategyType        string    `json:"strategy_type" validate:"required,oneof=savings investment expense_reduction"`
	InitialAmount       float64   `json:"initial_amount" validate:"required,gte=0"`
	MonthlyContribution float64   `json:"monthly_contribution" validate:"gte=0"`
	ExpectedReturn      float64   `json:"expected_return" validate:"gte=0,lte=100"`
	Duration            int       `json:"duration" validate:"required,gt=0"`
	StartDate           time.Time `json:"start_date" validate:"required"`
}

type UpdateBacktestStrategyRequest struct {
	Name                string   `json:"name" validate:"omitempty,min=1,max=255"`
	Description         string   `json:"description" validate:"max=1000"`
	StrategyType        string   `json:"strategy_type" validate:"omitempty,oneof=savings investment expense_reduction"`
	InitialAmount       *float64 `json:"initial_amount" validate:"omitempty,gte=0"`
	MonthlyContribution *float64 `json:"monthly_contribution" validate:"omitempty,gte=0"`
	ExpectedReturn      *float64 `json:"expected_return" validate:"omitempty,gte=0,lte=100"`
	Duration            *int     `json:"duration" validate:"omitempty,gt=0"`
	Status              string   `json:"status" validate:"omitempty,oneof=draft active completed cancelled"`
	ActualAmount        *float64 `json:"actual_amount" validate:"omitempty,gte=0"`
}

type BacktestStrategyResponse struct {
	ID                  uint      `json:"id"`
	UserID              string    `json:"user_id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	StrategyType        string    `json:"strategy_type"`
	InitialAmount       float64   `json:"initial_amount"`
	MonthlyContribution float64   `json:"monthly_contribution"`
	ExpectedReturn      float64   `json:"expected_return"`
	Duration            int       `json:"duration"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
	ProjectedAmount     float64   `json:"projected_amount"`
	ActualAmount        float64   `json:"actual_amount"`
	Status              string    `json:"status"`
	PerformanceGap      float64   `json:"performance_gap"`
	ResultData          string    `json:"result_data,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type BacktestStrategyListResponse struct {
	Strategies []BacktestStrategyResponse `json:"strategies"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}

func (b *BacktestStrategy) ToBacktestStrategyResponse() *BacktestStrategyResponse {
	response := &BacktestStrategyResponse{
		ID:                  b.ID,
		UserID:              b.UserID.String(),
		Name:                b.Name,
		Description:         b.Description,
		StrategyType:        b.StrategyType,
		InitialAmount:       b.InitialAmount,
		MonthlyContribution: b.MonthlyContribution,
		ExpectedReturn:      b.ExpectedReturn,
		Duration:            b.Duration,
		StartDate:           b.StartDate,
		EndDate:             b.EndDate,
		ProjectedAmount:     b.ProjectedAmount,
		ActualAmount:        b.ActualAmount,
		Status:              b.Status,
		ResultData:          b.ResultData,
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
	}

	// Calculate performance gap
	if b.ProjectedAmount > 0 {
		response.PerformanceGap = ((b.ActualAmount - b.ProjectedAmount) / b.ProjectedAmount) * 100
	}

	return response
}

// Budget DTOs
type CreateBudgetRequest struct {
	Category    string    `json:"category" validate:"required,min=1,max=100"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Period      string    `json:"period" validate:"required,oneof=weekly monthly yearly"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date"`
	IsRecurring bool      `json:"is_recurring"`
}

type UpdateBudgetRequest struct {
	Category    string     `json:"category" validate:"omitempty,min=1,max=100"`
	Amount      *float64   `json:"amount" validate:"omitempty,gt=0"`
	Period      string     `json:"period" validate:"omitempty,oneof=weekly monthly yearly"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	IsRecurring *bool      `json:"is_recurring"`
}

type BudgetResponse struct {
	ID          uint      `json:"id"`
	UserID      string    `json:"user_id"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Period      string    `json:"period"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date,omitempty"`
	SpentAmount float64   `json:"spent_amount"`
	Remaining   float64   `json:"remaining"`
	Progress    float64   `json:"progress"` // percentage
	IsRecurring bool      `json:"is_recurring"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BudgetListResponse struct {
	Budgets    []BudgetResponse `json:"budgets"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

func (b *Budget) ToBudgetResponse() *BudgetResponse {
	response := &BudgetResponse{
		ID:          b.ID,
		UserID:      b.UserID.String(),
		Category:    b.Category,
		Amount:      b.Amount,
		Period:      b.Period,
		StartDate:   b.StartDate,
		EndDate:     b.EndDate,
		SpentAmount: b.SpentAmount,
		IsRecurring: b.IsRecurring,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}

	// Calculate remaining and progress
	response.Remaining = b.Amount - b.SpentAmount
	if b.Amount > 0 {
		response.Progress = (b.SpentAmount / b.Amount) * 100
	}

	return response
}
