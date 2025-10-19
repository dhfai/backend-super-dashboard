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

type DailyTargetController struct {
	dailyTargetService *services.DailyTargetService
}

func NewDailyTargetController(dailyTargetService *services.DailyTargetService) *DailyTargetController {
	return &DailyTargetController{
		dailyTargetService: dailyTargetService,
	}
}

// CreateDailyTarget creates a new daily target
func (c *DailyTargetController) CreateDailyTarget(ctx *gin.Context) {
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

	var req models.CreateDailyTargetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	target, err := c.dailyTargetService.CreateDailyTarget(userID, &req)
	if err != nil {
		logger.GetLogger().Error("Failed to create daily target: ", err)

		// Handle specific error types
		if utils.IsAlreadyExistsError(err) {
			ctx.JSON(http.StatusConflict, models.APIResponse{
				Success: false,
				Message: "Daily target already exists for this date",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create daily target",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Daily target created successfully",
		Data:    target.ToDailyTargetResponse(),
	})
}

// GetDailyTargetByID retrieves a daily target by ID
func (c *DailyTargetController) GetDailyTargetByID(ctx *gin.Context) {
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

	target, err := c.dailyTargetService.GetDailyTargetByID(userID, uint(targetID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Daily target not found",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Daily target retrieved successfully",
		Data:    target.ToDailyTargetResponse(),
	})
}

// GetAllDailyTargets retrieves all daily targets
func (c *DailyTargetController) GetAllDailyTargets(ctx *gin.Context) {
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

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

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

	targets, total, err := c.dailyTargetService.GetAllDailyTargets(userID, page, pageSize, dateFrom, dateTo)
	if err != nil {
		logger.GetLogger().Error("Failed to get daily targets: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve daily targets",
			Error:   err.Error(),
		})
		return
	}

	targetResponses := make([]models.DailyTargetResponse, len(targets))
	for i, target := range targets {
		targetResponses[i] = *target.ToDailyTargetResponse()
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	response := models.DailyTargetListResponse{
		Targets:    targetResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Daily targets retrieved successfully",
		Data:    response,
	})
}

// UpdateDailyTarget updates a daily target
func (c *DailyTargetController) UpdateDailyTarget(ctx *gin.Context) {
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

	var req models.UpdateDailyTargetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	target, err := c.dailyTargetService.UpdateDailyTarget(userID, uint(targetID), &req)
	if err != nil {
		logger.GetLogger().Error("Failed to update daily target: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update daily target",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Daily target updated successfully",
		Data:    target.ToDailyTargetResponse(),
	})
}

// DeleteDailyTarget deletes a daily target
func (c *DailyTargetController) DeleteDailyTarget(ctx *gin.Context) {
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

	if err := c.dailyTargetService.DeleteDailyTarget(userID, uint(targetID)); err != nil {
		logger.GetLogger().Error("Failed to delete daily target: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete daily target",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Daily target deleted successfully",
	})
}

// GetTodayTarget retrieves today's target
func (c *DailyTargetController) GetTodayTarget(ctx *gin.Context) {
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

	target, err := c.dailyTargetService.GetTodayTarget(userID)
	if err != nil {
		logger.GetLogger().Error("Failed to get today's target: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve today's target",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Today's target retrieved successfully",
		Data:    target.ToDailyTargetResponse(),
	})
}

// GetCurrentMonthSummary retrieves current month summary
func (c *DailyTargetController) GetCurrentMonthSummary(ctx *gin.Context) {
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

	summary, err := c.dailyTargetService.GetCurrentMonthSummary(userID)
	if err != nil {
		logger.GetLogger().Error("Failed to get month summary: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve month summary",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Month summary retrieved successfully",
		Data:    summary,
	})
}

// GetWeekSummary retrieves current week summary
func (c *DailyTargetController) GetWeekSummary(ctx *gin.Context) {
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

	summary, err := c.dailyTargetService.GetWeekSummary(userID)
	if err != nil {
		logger.GetLogger().Error("Failed to get week summary: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve week summary",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Week summary retrieved successfully",
		Data:    summary,
	})
}

// RefreshActualValues refreshes actual values from transactions
func (c *DailyTargetController) RefreshActualValues(ctx *gin.Context) {
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

	target, err := c.dailyTargetService.RefreshActualValues(userID, uint(targetID))
	if err != nil {
		logger.GetLogger().Error("Failed to refresh actual values: ", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to refresh actual values",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Actual values refreshed successfully",
		Data:    target.ToDailyTargetResponse(),
	})
}
