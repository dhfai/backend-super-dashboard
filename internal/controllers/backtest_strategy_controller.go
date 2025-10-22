package controllers

import (
	"net/http"
	"strconv"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
)

type BacktestStrategyController struct {
	backtestService *services.BacktestStrategyService
}

func NewBacktestStrategyController(backtestService *services.BacktestStrategyService) *BacktestStrategyController {
	return &BacktestStrategyController{backtestService: backtestService}
}

func (c *BacktestStrategyController) CreateBacktestStrategy(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	var req models.CreateBacktestStrategyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	strategy, err := c.backtestService.CreateBacktestStrategy(userID, &req)
	if err != nil {
		logger.GetLogger().Error("Failed to create backtest strategy: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to create backtest strategy", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{Success: true, Message: "Backtest strategy created successfully", Data: strategy.ToBacktestStrategyResponse()})
}

func (c *BacktestStrategyController) GetBacktestStrategyByID(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	strategyID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	strategy, err := c.backtestService.GetBacktestStrategyByID(userID, uint(strategyID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "Backtest strategy not found", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Backtest strategy retrieved successfully", Data: strategy.ToBacktestStrategyResponse()})
}

func (c *BacktestStrategyController) GetAllBacktestStrategies(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	strategyType := ctx.Query("strategy_type")
	status := ctx.Query("status")

	strategies, total, err := c.backtestService.GetAllBacktestStrategies(userID, page, pageSize, strategyType, status)
	if err != nil {
		logger.GetLogger().Error("Failed to get backtest strategies: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to retrieve backtest strategies", Error: err.Error()})
		return
	}

	strategyResponses := make([]models.BacktestStrategyResponse, len(strategies))
	for i, strategy := range strategies {
		strategyResponses[i] = *strategy.ToBacktestStrategyResponse()
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	response := models.BacktestStrategyListResponse{Strategies: strategyResponses, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages}
	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Backtest strategies retrieved successfully", Data: response})
}

func (c *BacktestStrategyController) UpdateBacktestStrategy(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	strategyID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var req models.UpdateBacktestStrategyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	strategy, err := c.backtestService.UpdateBacktestStrategy(userID, uint(strategyID), &req)
	if err != nil {
		logger.GetLogger().Error("Failed to update backtest strategy: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to update backtest strategy", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Backtest strategy updated successfully", Data: strategy.ToBacktestStrategyResponse()})
}

func (c *BacktestStrategyController) DeleteBacktestStrategy(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	strategyID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err := c.backtestService.DeleteBacktestStrategy(userID, uint(strategyID)); err != nil {
		logger.GetLogger().Error("Failed to delete backtest strategy: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to delete backtest strategy", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Backtest strategy deleted successfully"})
}

func (c *BacktestStrategyController) RunBacktest(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	strategyID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	strategy, err := c.backtestService.RunBacktest(userID, uint(strategyID))
	if err != nil {
		logger.GetLogger().Error("Failed to run backtest: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to run backtest", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Backtest ran successfully", Data: strategy.ToBacktestStrategyResponse()})
}

func (c *BacktestStrategyController) GetBacktestResult(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	strategyID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	result, err := c.backtestService.GetBacktestResult(userID, uint(strategyID))
	if err != nil {
		logger.GetLogger().Error("Failed to get backtest result: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to retrieve backtest result", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Backtest result retrieved successfully", Data: result})
}

func (c *BacktestStrategyController) ActivateStrategy(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	strategyID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	strategy, err := c.backtestService.ActivateStrategy(userID, uint(strategyID))
	if err != nil {
		logger.GetLogger().Error("Failed to activate strategy: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to activate strategy", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Strategy activated successfully", Data: strategy.ToBacktestStrategyResponse()})
}

func (c *BacktestStrategyController) CompleteStrategy(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
		return
	}

	strategyID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var req struct {
		ActualAmount float64 `json:"actual_amount" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request data", Error: err.Error()})
		return
	}

	strategy, err := c.backtestService.CompleteStrategy(userID, uint(strategyID), req.ActualAmount)
	if err != nil {
		logger.GetLogger().Error("Failed to complete strategy: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to complete strategy", Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Strategy completed successfully", Data: strategy.ToBacktestStrategyResponse()})
}
