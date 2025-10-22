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

type BudgetController struct {
	budgetService *services.BudgetService
}

func NewBudgetController(budgetService *services.BudgetService) *BudgetController {
	return &BudgetController{budgetService: budgetService}
}

func (c *BudgetController) CreateBudget(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	var req models.CreateBudgetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	budget, err := c.budgetService.CreateBudget(userID, &req)
	if err != nil {
		logger.GetLogger().Error("Failed to create budget: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to create budget", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{Success: true, Message: "Budget created successfully", Data: budget.ToBudgetResponse()})
}

func (c *BudgetController) GetBudgetByID(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	budgetID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	budget, err := c.budgetService.GetBudgetByID(userID, uint(budgetID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "Budget not found", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Budget retrieved successfully", Data: budget.ToBudgetResponse()})
}

func (c *BudgetController) GetAllBudgets(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	category := ctx.Query("category")
	period := ctx.Query("period")

	var isActive *bool
	if ia := ctx.Query("is_active"); ia != "" {
		iaVal := ia == "true"
		isActive = &iaVal
	}

	budgets, total, err := c.budgetService.GetAllBudgets(userID, page, pageSize, category, period, isActive)
	if err != nil {
		logger.GetLogger().Error("Failed to get budgets: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to retrieve budgets", Error: err.Error()})
		return
	}

	budgetResponses := make([]models.BudgetResponse, len(budgets))
	for i, budget := range budgets {
		budgetResponses[i] = *budget.ToBudgetResponse()
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	response := models.BudgetListResponse{Budgets: budgetResponses, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages}
	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Budgets retrieved successfully", Data: response})
}

func (c *BudgetController) UpdateBudget(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	budgetID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var req models.UpdateBudgetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	budget, err := c.budgetService.UpdateBudget(userID, uint(budgetID), &req)
	if err != nil {
		logger.GetLogger().Error("Failed to update budget: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to update budget", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Budget updated successfully", Data: budget.ToBudgetResponse()})
}

func (c *BudgetController) DeleteBudget(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	budgetID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err := c.budgetService.DeleteBudget(userID, uint(budgetID)); err != nil {
		logger.GetLogger().Error("Failed to delete budget: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to delete budget", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Budget deleted successfully"})
}

func (c *BudgetController) GetBudgetSummary(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	summary, err := c.budgetService.GetBudgetSummary(userID)
	if err != nil {
		logger.GetLogger().Error("Failed to get budget summary: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to retrieve budget summary", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Budget summary retrieved successfully", Data: summary})
}

func (c *BudgetController) GetMonthlyBudgetReport(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	now := time.Now()
	year, _ := strconv.Atoi(ctx.DefaultQuery("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(ctx.DefaultQuery("month", strconv.Itoa(int(now.Month()))))

	report, err := c.budgetService.GetMonthlyBudgetReport(userID, year, month)
	if err != nil {
		logger.GetLogger().Error("Failed to get monthly budget report: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to retrieve monthly budget report", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Monthly budget report retrieved successfully", Data: report})
}

func (c *BudgetController) CheckBudgetAlert(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	threshold, _ := strconv.ParseFloat(ctx.DefaultQuery("threshold", "80"), 64)
	alerts, err := c.budgetService.CheckBudgetAlert(userID, threshold)
	if err != nil {
		logger.GetLogger().Error("Failed to check budget alerts: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to check budget alerts", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Budget alerts retrieved successfully", Data: alerts})
}
