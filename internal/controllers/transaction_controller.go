package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	transactionService *services.TransactionService
}

func NewTransactionController(transactionService *services.TransactionService) *TransactionController {
	return &TransactionController{
		transactionService: transactionService,
	}
}

// CreateTransaction creates a new transaction
// @Summary Create transaction
// @Description Create a new financial transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Param transaction body models.CreateTransactionRequest true "Transaction data"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Router /api/v1/finance/transactions [post]
func (c *TransactionController) CreateTransaction(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	var req models.CreateTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	transaction, err := c.transactionService.CreateTransaction(userID, &req)
	if err != nil {
		logger.GetLogger().Error("Failed to create transaction: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create transaction",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Transaction created successfully",
		Data:    transaction.ToTransactionResponse(),
	})
}

// GetTransactionByID retrieves a transaction by ID
// @Summary Get transaction
// @Description Get a transaction by ID
// @Tags Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /api/v1/finance/transactions/{id} [get]
func (c *TransactionController) GetTransactionByID(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	transactionID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid transaction ID",
			Error:   err.Error(),
		})
		return
	}

	transaction, err := c.transactionService.GetTransactionByID(userID, uint(transactionID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Transaction not found",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Transaction retrieved successfully",
		Data:    transaction.ToTransactionResponse(),
	})
}

// GetAllTransactions retrieves all transactions with filters
// @Summary Get all transactions
// @Description Get all transactions with optional filters
// @Tags Transactions
// @Produce json
// @Param type query string false "Transaction type (income/expense)"
// @Param category query string false "Category"
// @Param tags query string false "Tags"
// @Param date_from query string false "Date from (YYYY-MM-DD)"
// @Param date_to query string false "Date to (YYYY-MM-DD)"
// @Param min_amount query number false "Minimum amount"
// @Param max_amount query number false "Maximum amount"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param sort_by query string false "Sort by (date/amount/created_at)"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/finance/transactions [get]
func (c *TransactionController) GetAllTransactions(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	var filter models.TransactionFilterRequest
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid filter parameters",
			Error:   err.Error(),
		})
		return
	}

	// Parse date strings if provided
	if dateFromStr := ctx.Query("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err == nil {
			filter.DateFrom = &dateFrom
		}
	}
	if dateToStr := ctx.Query("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err == nil {
			filter.DateTo = &dateTo
		}
	}

	transactions, total, err := c.transactionService.GetAllTransactions(userID, &filter)
	if err != nil {
		logger.GetLogger().Error("Failed to get transactions: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve transactions",
			Error:   err.Error(),
		})
		return
	}

	// Convert to response
	transactionResponses := make([]models.TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		transactionResponses[i] = *transaction.ToTransactionResponse()
	}

	// Get summary
	summary, _ := c.transactionService.GetTransactionSummary(userID, filter.DateFrom, filter.DateTo)

	// Calculate pagination
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize != 0 {
		totalPages++
	}

	response := models.TransactionListResponse{
		Transactions: transactionResponses,
		Total:        total,
		Page:         filter.Page,
		PageSize:     filter.PageSize,
		TotalPages:   totalPages,
		Summary:      summary,
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Transactions retrieved successfully",
		Data:    response,
	})
}

// UpdateTransaction updates a transaction
// @Summary Update transaction
// @Description Update an existing transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param transaction body models.UpdateTransactionRequest true "Transaction data"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Router /api/v1/finance/transactions/{id} [put]
func (c *TransactionController) UpdateTransaction(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	transactionID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid transaction ID",
			Error:   err.Error(),
		})
		return
	}

	var req models.UpdateTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	transaction, err := c.transactionService.UpdateTransaction(userID, uint(transactionID), &req)
	if err != nil {
		logger.GetLogger().Error("Failed to update transaction: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update transaction",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Transaction updated successfully",
		Data:    transaction.ToTransactionResponse(),
	})
}

// DeleteTransaction soft deletes a transaction
// @Summary Delete transaction
// @Description Soft delete a transaction
// @Tags Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/finance/transactions/{id} [delete]
func (c *TransactionController) DeleteTransaction(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	transactionID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid transaction ID",
			Error:   err.Error(),
		})
		return
	}

	if err := c.transactionService.DeleteTransaction(userID, uint(transactionID)); err != nil {
		logger.GetLogger().Error("Failed to delete transaction: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete transaction",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Transaction deleted successfully",
	})
}

// GetTransactionsByCategory retrieves transactions by category
// @Summary Get transactions by category
// @Description Get transactions filtered by category
// @Tags Transactions
// @Produce json
// @Param category path string true "Category"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/finance/transactions/category/{category} [get]
func (c *TransactionController) GetTransactionsByCategory(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	category := ctx.Param("category")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	transactions, total, err := c.transactionService.GetTransactionsByCategory(userID, category, page, pageSize)
	if err != nil {
		logger.GetLogger().Error("Failed to get transactions by category: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve transactions",
			Error:   err.Error(),
		})
		return
	}

	transactionResponses := make([]models.TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		transactionResponses[i] = *transaction.ToTransactionResponse()
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	response := models.TransactionListResponse{
		Transactions: transactionResponses,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Transactions retrieved successfully",
		Data:    response,
	})
}

// GetTransactionSummary retrieves transaction summary
// @Summary Get transaction summary
// @Description Get summary of income, expense, and balance
// @Tags Transactions
// @Produce json
// @Param date_from query string false "Date from (YYYY-MM-DD)"
// @Param date_to query string false "Date to (YYYY-MM-DD)"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/finance/transactions/summary [get]
func (c *TransactionController) GetTransactionSummary(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	var dateFrom, dateTo *time.Time
	if dateFromStr := ctx.Query("date_from"); dateFromStr != "" {
		df, err := time.Parse("2006-01-02", dateFromStr)
		if err == nil {
			dateFrom = &df
		}
	}
	if dateToStr := ctx.Query("date_to"); dateToStr != "" {
		dt, err := time.Parse("2006-01-02", dateToStr)
		if err == nil {
			dateTo = &dt
		}
	}

	summary, err := c.transactionService.GetTransactionSummary(userID, dateFrom, dateTo)
	if err != nil {
		logger.GetLogger().Error("Failed to get transaction summary: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve summary",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Summary retrieved successfully",
		Data:    summary,
	})
}

// GetCategoryBreakdown retrieves category breakdown
// @Summary Get category breakdown
// @Description Get spending breakdown by category
// @Tags Transactions
// @Produce json
// @Param type query string false "Transaction type (income/expense)"
// @Param date_from query string false "Date from (YYYY-MM-DD)"
// @Param date_to query string false "Date to (YYYY-MM-DD)"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/finance/transactions/breakdown [get]
func (c *TransactionController) GetCategoryBreakdown(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	transactionType := ctx.Query("type")
	var dateFrom, dateTo *time.Time
	if dateFromStr := ctx.Query("date_from"); dateFromStr != "" {
		df, err := time.Parse("2006-01-02", dateFromStr)
		if err == nil {
			dateFrom = &df
		}
	}
	if dateToStr := ctx.Query("date_to"); dateToStr != "" {
		dt, err := time.Parse("2006-01-02", dateToStr)
		if err == nil {
			dateTo = &dt
		}
	}

	breakdown, err := c.transactionService.GetCategoryBreakdown(userID, transactionType, dateFrom, dateTo)
	if err != nil {
		logger.GetLogger().Error("Failed to get category breakdown: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve breakdown",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Breakdown retrieved successfully",
		Data:    breakdown,
	})
}

// GetMonthlyTrend retrieves monthly trend
// @Summary Get monthly trend
// @Description Get monthly transaction trends
// @Tags Transactions
// @Produce json
// @Param months query int false "Number of months (default 6)"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/finance/transactions/trend [get]
func (c *TransactionController) GetMonthlyTrend(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.GetLogger().Error("Failed to get user ID from context: ", err)
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
		return
	}

	months, _ := strconv.Atoi(ctx.DefaultQuery("months", "6"))
	if months < 1 {
		months = 6
	}

	trend, err := c.transactionService.GetMonthlyTrend(userID, months)
	if err != nil {
		logger.GetLogger().Error("Failed to get monthly trend: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve trend",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trend retrieved successfully",
		Data:    trend,
	})
}
