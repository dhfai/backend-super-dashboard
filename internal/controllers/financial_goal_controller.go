package controllers

import (
	"net/http"
	"strconv"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
)

type FinancialGoalController struct {
	financialGoalService *services.FinancialGoalService
}

func NewFinancialGoalController(financialGoalService *services.FinancialGoalService) *FinancialGoalController {
	return &FinancialGoalController{
		financialGoalService: financialGoalService,
	}
}

func (c *FinancialGoalController) CreateFinancialGoal(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	var req models.CreateFinancialGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	goal, err := c.financialGoalService.CreateFinancialGoal(userID, &req)
	if err != nil {
		logger.GetLogger().Error("Failed to create financial goal: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to create financial goal", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{Success: true, Message: "Financial goal created successfully", Data: goal.ToFinancialGoalResponse()})
}

func (c *FinancialGoalController) GetFinancialGoalByID(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	goalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	goal, err := c.financialGoalService.GetFinancialGoalByID(userID, uint(goalID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "Financial goal not found", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Financial goal retrieved successfully", Data: goal.ToFinancialGoalResponse()})
}

func (c *FinancialGoalController) GetAllFinancialGoals(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	category := ctx.Query("category")

	var priority *int
	if p := ctx.Query("priority"); p != "" {
		pVal, _ := strconv.Atoi(p)
		priority = &pVal
	}

	var isCompleted *bool
	if ic := ctx.Query("is_completed"); ic != "" {
		icVal := ic == "true"
		isCompleted = &icVal
	}

	goals, total, err := c.financialGoalService.GetAllFinancialGoals(userID, page, pageSize, category, priority, isCompleted)
	if err != nil {
		logger.GetLogger().Error("Failed to get financial goals: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to retrieve financial goals", Error: err.Error()})
		return
	}

	goalResponses := make([]models.FinancialGoalResponse, len(goals))
	for i, goal := range goals {
		goalResponses[i] = *goal.ToFinancialGoalResponse()
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	response := models.FinancialGoalListResponse{Goals: goalResponses, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages}
	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Financial goals retrieved successfully", Data: response})
}

func (c *FinancialGoalController) UpdateFinancialGoal(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	goalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var req models.UpdateFinancialGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	goal, err := c.financialGoalService.UpdateFinancialGoal(userID, uint(goalID), &req)
	if err != nil {
		logger.GetLogger().Error("Failed to update financial goal: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to update financial goal", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Financial goal updated successfully", Data: goal.ToFinancialGoalResponse()})
}

func (c *FinancialGoalController) DeleteFinancialGoal(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	goalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err := c.financialGoalService.DeleteFinancialGoal(userID, uint(goalID)); err != nil {
		logger.GetLogger().Error("Failed to delete financial goal: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to delete financial goal", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Financial goal deleted successfully"})
}

func (c *FinancialGoalController) UpdateGoalProgress(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	goalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	goal, err := c.financialGoalService.UpdateGoalProgress(userID, uint(goalID), req.Amount)
	if err != nil {
		logger.GetLogger().Error("Failed to update goal progress: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to update goal progress", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Goal progress updated successfully", Data: goal.ToFinancialGoalResponse()})
}

func (c *FinancialGoalController) AddToGoalProgress(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	goalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	goal, err := c.financialGoalService.AddToGoalProgress(userID, uint(goalID), req.Amount)
	if err != nil {
		logger.GetLogger().Error("Failed to add to goal progress: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to add to goal progress", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Amount added to goal successfully", Data: goal.ToFinancialGoalResponse()})
}

func (c *FinancialGoalController) GetGoalsSummary(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	summary, err := c.financialGoalService.GetGoalsSummary(userID)
	if err != nil {
		logger.GetLogger().Error("Failed to get goals summary: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to retrieve goals summary", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Goals summary retrieved successfully", Data: summary})
}
