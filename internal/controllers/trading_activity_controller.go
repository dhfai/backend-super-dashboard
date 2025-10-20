package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/dhfai/dhfai-gobackend.git/internal/utils"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
)

type TradingActivityController struct {
	tradingActivityService *services.TradingActivityService
}

func NewTradingActivityController(tradingActivityService *services.TradingActivityService) *TradingActivityController {
	return &TradingActivityController{
		tradingActivityService: tradingActivityService,
	}
}

// AddTradeToTarget adds a trading activity to a daily target
func (c *TradingActivityController) AddTradeToTarget(ctx *gin.Context) {
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

	targetID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid target ID",
			Error:   err.Error(),
		})
		return
	}

	var req models.CreateTradingActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	activity, target, err := c.tradingActivityService.AddTradeToTarget(userID, uint(targetID), &req)
	if err != nil {
		logger.GetLogger().Error("Failed to add trade: ", err)

		// Handle specific error types
		if utils.IsNotFoundError(err) {
			ctx.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "Daily target not found",
				Error:   err.Error(),
			})
			return
		}

		if utils.IsInvalidInputError(err) {
			ctx.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Invalid input",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to add trade",
			Error:   err.Error(),
		})
		return
	}

	// Prepare response with both activity and updated target
	response := map[string]interface{}{
		"activity": activity.ToTradingActivityResponse(),
		"target":   target.ToDailyTargetResponse(),
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Trade added successfully",
		Data:    response,
	})
}

// GetTradingActivitiesByTarget retrieves all trading activities for a specific daily target
func (c *TradingActivityController) GetTradingActivitiesByTarget(ctx *gin.Context) {
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

	targetID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid target ID",
			Error:   err.Error(),
		})
		return
	}

	activities, err := c.tradingActivityService.GetTradingActivitiesByTarget(userID, uint(targetID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve trading activities",
			Error:   err.Error(),
		})
		return
	}

	// Convert to response
	activityResponses := make([]models.TradingActivityResponse, len(activities))
	for i, activity := range activities {
		activityResponses[i] = *activity.ToTradingActivityResponse()
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trading activities retrieved successfully",
		Data:    activityResponses,
	})
}

// GetTradingActivityByID retrieves a specific trading activity
func (c *TradingActivityController) GetTradingActivityByID(ctx *gin.Context) {
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

	activityID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid activity ID",
			Error:   err.Error(),
		})
		return
	}

	activity, err := c.tradingActivityService.GetTradingActivityByID(userID, uint(activityID))
	if err != nil {
		if utils.IsNotFoundError(err) {
			ctx.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "Trading activity not found",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve trading activity",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trading activity retrieved successfully",
		Data:    activity.ToTradingActivityResponse(),
	})
}

// DeleteTradingActivity deletes a trading activity
func (c *TradingActivityController) DeleteTradingActivity(ctx *gin.Context) {
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

	activityID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid activity ID",
			Error:   err.Error(),
		})
		return
	}

	if err := c.tradingActivityService.DeleteTradingActivity(userID, uint(activityID)); err != nil {
		if utils.IsNotFoundError(err) {
			ctx.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "Trading activity not found",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete trading activity",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trading activity deleted successfully",
	})
}

// GetTradingStats retrieves trading statistics
func (c *TradingActivityController) GetTradingStats(ctx *gin.Context) {
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

	// Parse date filters
	var dateFrom, dateTo time.Time
	if dateFromStr := ctx.Query("date_from"); dateFromStr != "" {
		if df, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = df
		}
	}
	if dateToStr := ctx.Query("date_to"); dateToStr != "" {
		if dt, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = dt
		}
	}

	stats, err := c.tradingActivityService.GetTradingStats(userID, dateFrom, dateTo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve trading stats",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trading stats retrieved successfully",
		Data:    stats,
	})
}
