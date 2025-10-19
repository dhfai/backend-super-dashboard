package controllers

import (
	"net/http"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/utils"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProfileController handles profile related requests
type ProfileController struct {
	db *gorm.DB
}

// NewProfileController creates a new ProfileController
func NewProfileController(db *gorm.DB) *ProfileController {
	return &ProfileController{
		db: db,
	}
}

// GetProfile retrieves user profile
func (pc *ProfileController) GetProfile(c *gin.Context) {
	log := logger.GetLogger()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		log.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	// Find user with profile
	var user models.User
	if err := pc.db.Preload("Profile").Where("id = ?", userUUID).First(&user).Error; err != nil {
		log.WithError(err).WithField("user_id", userUUID).Error("Failed to find user")
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve profile",
		})
		return
	}

	log.WithField("user_id", userUUID).Info("Profile retrieved successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    user.ToUserResponse(),
	})
}

// UpdateProfile updates user profile
func (pc *ProfileController) UpdateProfile(c *gin.Context) {
	log := logger.GetLogger()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		log.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind update profile request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Profile update validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	// Sanitize input data
	req.FullName = utils.SanitizeString(req.FullName)
	req.Address = utils.SanitizeString(req.Address)
	req.PhoneNumber = utils.SanitizeString(req.PhoneNumber)
	req.Country = utils.SanitizeString(req.Country)

	// Find existing profile or create new one
	var profile models.Profile
	err = pc.db.Where("user_id = ?", userUUID).First(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new profile
			profile = models.Profile{
				UserID:      userUUID,
				FullName:    req.FullName,
				Address:     req.Address,
				PhoneNumber: req.PhoneNumber,
				Country:     req.Country,
			}

			if err := pc.db.Create(&profile).Error; err != nil {
				log.WithError(err).Error("Failed to create profile")
				c.JSON(http.StatusInternalServerError, models.APIResponse{
					Success: false,
					Message: "Failed to create profile",
				})
				return
			}
		} else {
			log.WithError(err).Error("Failed to find profile")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to update profile",
			})
			return
		}
	} else {
		// Update existing profile
		profile.FullName = req.FullName
		profile.Address = req.Address
		profile.PhoneNumber = req.PhoneNumber
		profile.Country = req.Country

		if err := pc.db.Save(&profile).Error; err != nil {
			log.WithError(err).Error("Failed to update profile")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to update profile",
			})
			return
		}
	}

	// Get updated user with profile
	var user models.User
	if err := pc.db.Preload("Profile").Where("id = ?", userUUID).First(&user).Error; err != nil {
		log.WithError(err).Error("Failed to get updated user")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Profile updated but failed to retrieve updated data",
		})
		return
	}

	log.WithField("user_id", userUUID).Info("Profile updated successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data:    user.ToUserResponse(),
	})
}

// DeleteProfile deletes user profile (soft delete)
func (pc *ProfileController) DeleteProfile(c *gin.Context) {
	log := logger.GetLogger()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		log.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	// Find and delete profile
	var profile models.Profile
	if err := pc.db.Where("user_id = ?", userUUID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WithField("user_id", userUUID).Warn("Profile not found for deletion")
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "Profile not found",
			})
			return
		}
		log.WithError(err).Error("Failed to find profile")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete profile",
		})
		return
	}

	// Delete profile (GORM soft delete)
	if err := pc.db.Delete(&profile).Error; err != nil {
		log.WithError(err).Error("Failed to delete profile")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete profile",
		})
		return
	}

	log.WithField("user_id", userUUID).Info("Profile deleted successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile deleted successfully",
	})
}

// GetUserInfo retrieves basic user information (without profile)
func (pc *ProfileController) GetUserInfo(c *gin.Context) {
	log := logger.GetLogger()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		log.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	// Find user
	var user models.User
	if err := pc.db.Where("id = ?", userUUID).First(&user).Error; err != nil {
		log.WithError(err).WithField("user_id", userUUID).Error("Failed to find user")
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve user information",
		})
		return
	}

	log.WithField("user_id", userUUID).Info("User info retrieved successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User information retrieved successfully",
		Data:    user.ToUserResponse(),
	})
}
