package controllers

import (
	"net/http"
	"time"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/dhfai/dhfai-gobackend.git/internal/utils"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	db           *gorm.DB
	authService  *services.AuthService
	emailService *services.EmailService
}

func NewAuthController(db *gorm.DB, authService *services.AuthService, emailService *services.EmailService) *AuthController {
	return &AuthController{
		db:           db,
		authService:  authService,
		emailService: emailService,
	}
}

func (ac *AuthController) Register(c *gin.Context) {
	log := logger.GetLogger()

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind registration request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Registration validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)
	req.Username = utils.NormalizeUsername(req.Username)

	// Check if user already exists
	var existingUser models.User
	if err := ac.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		log.WithField("email", req.Email).WithField("username", req.Username).Warn("User already exists")
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Message: "User with this email or username already exists",
		})
		return
	}

	// Check if temp registration already exists
	var existingTempReg models.TempRegistration
	if err := ac.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingTempReg).Error; err == nil {
		// Delete existing temp registration to allow re-registration
		ac.db.Delete(&existingTempReg)
		// Also delete related OTPs
		ac.db.Where("temp_registration_id = ?", existingTempReg.ID).Delete(&models.OTP{})
	}

	hashedPassword, err := ac.authService.HashPassword(req.Password)
	if err != nil {
		log.WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process registration",
		})
		return
	}

	// Save to temporary registration table
	tempReg := models.TempRegistration{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours to complete registration
	}

	if err := ac.db.Create(&tempReg).Error; err != nil {
		log.WithError(err).Error("Failed to create temporary registration")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process registration",
		})
		return
	}

	// Generate OTP for registration verification
	otpCode, err := ac.authService.GenerateOTP()
	if err != nil {
		log.WithError(err).Error("Failed to generate registration OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process registration",
		})
		return
	}

	// Save OTP linked to temp registration
	otp := models.OTP{
		TempRegistrationID: &tempReg.ID,
		Code:               otpCode,
		Type:               "register_verify",
		ExpiresAt:          time.Now().Add(15 * time.Minute),
		Used:               false,
	}

	if err := ac.db.Create(&otp).Error; err != nil {
		log.WithError(err).Error("Failed to save registration OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process registration",
		})
		return
	}

	// Send OTP email
	if err := ac.emailService.SendOTPEmail(tempReg.Email, otpCode, "register_verify"); err != nil {
		log.WithError(err).Error("Failed to send registration verification email")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to send verification email",
		})
		return
	}

	log.WithField("temp_reg_id", tempReg.ID).WithField("email", tempReg.Email).Info("Registration initiated, verification email sent")

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Registration initiated successfully. Please check your email for verification code to complete registration.",
		Data: map[string]interface{}{
			"email":      tempReg.Email,
			"expires_in": "24 hours",
		},
	})
}

func (ac *AuthController) VerifyRegistration(c *gin.Context) {
	log := logger.GetLogger()

	var req models.VerifyRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind verify registration request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Verify registration validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	// Find temp registration
	var tempReg models.TempRegistration
	if err := ac.db.Where("email = ?", req.Email).First(&tempReg).Error; err != nil {
		log.WithField("email", req.Email).Warn("Temp registration not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Registration not found or expired",
		})
		return
	}

	// Check if temp registration is expired
	if tempReg.ExpiresAt.Before(time.Now()) {
		log.WithField("temp_reg_id", tempReg.ID).Warn("Temp registration expired")
		// Clean up expired temp registration
		ac.db.Delete(&tempReg)
		ac.db.Where("temp_registration_id = ?", tempReg.ID).Delete(&models.OTP{})
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Registration expired. Please register again.",
		})
		return
	}

	// Find and verify OTP
	var otp models.OTP
	if err := ac.db.Where("temp_registration_id = ? AND code = ? AND type = ? AND used = ?",
		tempReg.ID, req.OTPCode, "register_verify", false).First(&otp).Error; err != nil {
		log.WithField("temp_reg_id", tempReg.ID).WithField("code", req.OTPCode).Warn("Invalid OTP")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid or expired OTP code",
		})
		return
	}

	// Check if OTP is expired
	if otp.ExpiresAt.Before(time.Now()) {
		log.WithField("otp_id", otp.ID).Warn("OTP expired")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "OTP code has expired. Please request a new one.",
		})
		return
	}

	// Begin transaction
	tx := ac.db.Begin()
	if tx.Error != nil {
		log.WithError(tx.Error).Error("Failed to begin transaction")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process registration",
		})
		return
	}

	// Create user from temp registration
	user := models.User{
		Username:        tempReg.Username,
		Email:           tempReg.Email,
		Password:        tempReg.Password,
		IsActive:        true,
		EmailVerified:   true,
		EmailVerifiedAt: &[]time.Time{time.Now()}[0],
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to create user")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to complete registration",
		})
		return
	}

	// Create user profile
	profile := models.Profile{
		UserID: user.ID,
	}
	if err := tx.Create(&profile).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to create user profile")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to complete registration",
		})
		return
	}

	// Mark OTP as used
	otp.Used = true
	if err := tx.Save(&otp).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to mark OTP as used")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to complete registration",
		})
		return
	}

	// Delete temp registration and related OTPs
	if err := tx.Delete(&tempReg).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to delete temp registration")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to complete registration",
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.WithError(err).Error("Failed to commit transaction")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to complete registration",
		})
		return
	}

	log.WithField("user_id", user.ID).WithField("email", user.Email).Info("Registration completed successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Registration completed successfully. You can now login.",
		Data:    user.ToUserResponse(),
	})
}

func (ac *AuthController) ResendRegistrationOTP(c *gin.Context) {
	log := logger.GetLogger()

	var req models.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind resend registration OTP request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Resend registration OTP validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	// Find temp registration
	var tempReg models.TempRegistration
	if err := ac.db.Where("email = ?", req.Email).First(&tempReg).Error; err != nil {
		log.WithField("email", req.Email).Warn("Temp registration not found for resend OTP")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Registration not found or expired",
		})
		return
	}

	// Check if temp registration is expired
	if tempReg.ExpiresAt.Before(time.Now()) {
		log.WithField("temp_reg_id", tempReg.ID).Warn("Temp registration expired for resend OTP")
		// Clean up expired temp registration
		ac.db.Delete(&tempReg)
		ac.db.Where("temp_registration_id = ?", tempReg.ID).Delete(&models.OTP{})
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Registration expired. Please register again.",
		})
		return
	}

	// Delete existing unused OTPs
	ac.db.Where("temp_registration_id = ? AND type = ? AND used = ?",
		tempReg.ID, "register_verify", false).Delete(&models.OTP{})

	// Generate new OTP
	otpCode, err := ac.authService.GenerateOTP()
	if err != nil {
		log.WithError(err).Error("Failed to generate new registration OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	// Save new OTP
	otp := models.OTP{
		TempRegistrationID: &tempReg.ID,
		Code:               otpCode,
		Type:               "register_verify",
		ExpiresAt:          time.Now().Add(15 * time.Minute),
		Used:               false,
	}

	if err := ac.db.Create(&otp).Error; err != nil {
		log.WithError(err).Error("Failed to save new registration OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	// Send OTP email
	if err := ac.emailService.SendOTPEmail(tempReg.Email, otpCode, "register_verify"); err != nil {
		log.WithError(err).Error("Failed to send new registration verification email")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to send verification email",
		})
		return
	}

	log.WithField("temp_reg_id", tempReg.ID).WithField("email", tempReg.Email).Info("Registration verification OTP resent")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "New verification code sent to your email.",
	})
}

func (ac *AuthController) Login(c *gin.Context) {
	log := logger.GetLogger()

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind login request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Login validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	var user models.User
	if err := ac.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		log.WithField("email", req.Email).Warn("User not found")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	if !user.IsActive {
		log.WithField("user_id", user.ID).Warn("Inactive user attempted login")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Account is inactive",
		})
		return
	}

	if !user.EmailVerified {
		log.WithField("user_id", user.ID).Warn("Unverified email attempted login")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Please verify your email before logging in. Check your inbox for verification code.",
		})
		return
	}

	if !ac.authService.VerifyPassword(user.Password, req.Password) {
		log.WithField("user_id", user.ID).Warn("Invalid password attempt")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	token, err := ac.authService.GenerateToken(user.ID.String(), user.Email)
	if err != nil {
		log.WithError(err).Error("Failed to generate token")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate authentication token",
		})
		return
	}

	var profile models.Profile
	if err := ac.db.Where("user_id = ?", user.ID).First(&profile).Error; err == nil {
		user.Profile = &profile
	}

	log.WithField("user_id", user.ID).Info("User logged in successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data: models.LoginResponse{
			Token: token,
			User:  user.ToUserResponse(),
		},
	})
}

func (ac *AuthController) ForgetPassword(c *gin.Context) {
	log := logger.GetLogger()

	var req models.ForgetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind forget password request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Forget password validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	var user models.User
	if err := ac.db.Where("email = ?", req.Email).First(&user).Error; err != nil {

		log.WithField("email", req.Email).Warn("Forget password request for non-existent email")
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "If your email is registered, you will receive a password reset code",
		})
		return
	}

	otpCode, err := ac.authService.GenerateOTP()
	if err != nil {
		log.WithError(err).Error("Failed to generate OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	otp := models.OTP{
		UserID:    &user.ID,
		Code:      otpCode,
		Type:      "reset_password",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	if err := ac.db.Create(&otp).Error; err != nil {
		log.WithError(err).Error("Failed to save OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	if err := ac.emailService.SendOTPEmail(user.Email, otpCode, "reset_password"); err != nil {
		log.WithError(err).Error("Failed to send OTP email")

	}

	log.WithField("user_id", user.ID).Info("Password reset OTP generated")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password reset code sent to your email",
	})
}

func (ac *AuthController) ResetPassword(c *gin.Context) {
	log := logger.GetLogger()

	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind reset password request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Reset password validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	var user models.User
	if err := ac.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		log.WithField("email", req.Email).Warn("Reset password request for non-existent email")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid email or OTP code",
		})
		return
	}

	var otp models.OTP
	if err := ac.db.Where("user_id = ? AND code = ? AND type = ? AND used = ? AND expires_at > ?",
		user.ID, req.OTPCode, "reset_password", false, time.Now()).First(&otp).Error; err != nil {
		log.WithField("user_id", user.ID).WithField("otp_code", req.OTPCode).Warn("Invalid or expired OTP")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid or expired OTP code",
		})
		return
	}

	hashedPassword, err := ac.authService.HashPassword(req.NewPassword)
	if err != nil {
		log.WithError(err).Error("Failed to hash new password")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	tx := ac.db.Begin()

	if err := tx.Model(&user).Update("password", hashedPassword).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to update password")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update password",
		})
		return
	}

	if err := tx.Model(&otp).Update("used", true).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to mark OTP as used")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	tx.Commit()

	log.WithField("user_id", user.ID).Info("Password reset successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}

func (ac *AuthController) VerifyEmail(c *gin.Context) {
	log := logger.GetLogger()

	var req models.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind verify email request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Email verification validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	var user models.User
	if err := ac.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		log.WithField("email", req.Email).Warn("Email verification request for non-existent email")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid email or verification code",
		})
		return
	}

	if user.EmailVerified {
		log.WithField("user_id", user.ID).Warn("Email already verified")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Email is already verified",
		})
		return
	}

	var otp models.OTP
	if err := ac.db.Where("user_id = ? AND code = ? AND type = ? AND used = ? AND expires_at > ?",
		user.ID, req.OTPCode, "verify_email", false, time.Now()).First(&otp).Error; err != nil {
		log.WithField("user_id", user.ID).WithField("otp_code", req.OTPCode).Warn("Invalid or expired verification OTP")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid or expired verification code",
		})
		return
	}

	tx := ac.db.Begin()

	now := time.Now()

	if err := tx.Model(&user).Updates(map[string]interface{}{
		"email_verified":    true,
		"email_verified_at": &now,
	}).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to update email verification status")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to verify email",
		})
		return
	}

	if err := tx.Model(&otp).Update("used", true).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to mark OTP as used")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process verification",
		})
		return
	}

	tx.Commit()

	log.WithField("user_id", user.ID).Info("Email verified successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Email verified successfully. You can now login.",
	})
}

func (ac *AuthController) ResendVerification(c *gin.Context) {
	log := logger.GetLogger()

	var req models.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind resend verification request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Resend verification validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	var user models.User
	if err := ac.db.Where("email = ?", req.Email).First(&user).Error; err != nil {

		log.WithField("email", req.Email).Warn("Resend verification request for non-existent email")
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "If your email is registered and not verified, you will receive a verification code",
		})
		return
	}

	if user.EmailVerified {
		log.WithField("user_id", user.ID).Warn("Resend verification for already verified email")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Email is already verified",
		})
		return
	}

	otpCode, err := ac.authService.GenerateOTP()
	if err != nil {
		log.WithError(err).Error("Failed to generate verification OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	otp := models.OTP{
		UserID:    &user.ID,
		Code:      otpCode,
		Type:      "verify_email",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	if err := ac.db.Create(&otp).Error; err != nil {
		log.WithError(err).Error("Failed to save verification OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	if err := ac.emailService.SendOTPEmail(user.Email, otpCode, "verify_email"); err != nil {
		log.WithError(err).Error("Failed to send verification email")

	}

	log.WithField("user_id", user.ID).Info("Email verification OTP resent")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Verification code sent to your email",
	})
}

func (ac *AuthController) Logout(c *gin.Context) {
	log := logger.GetLogger()

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "No token provided",
		})
		return
	}

	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	expiresAt, err := ac.authService.ExtractTokenExpiry(tokenString)
	if err != nil {
		log.WithError(err).Error("Failed to extract token expiry")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid token",
		})
		return
	}

	blacklistEntry := models.TokenBlacklist{
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}

	if err := ac.db.Create(&blacklistEntry).Error; err != nil {
		log.WithError(err).Error("Failed to blacklist token")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to logout",
		})
		return
	}

	log.Info("User successfully logged out")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Successfully logged out",
	})
}

func (ac *AuthController) RequestDeleteAccount(c *gin.Context) {
	log := logger.GetLogger()

	var req models.RequestDeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind delete account request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Delete account validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	var user models.User
	if err := ac.db.Where("id = ?", userID).First(&user).Error; err != nil {
		log.WithError(err).Error("User not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	if !ac.authService.VerifyPassword(user.Password, req.Password) {
		log.WithField("user_id", user.ID).Warn("Invalid password for delete account request")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid password",
		})
		return
	}

	otpCode, err := ac.authService.GenerateOTP()
	if err != nil {
		log.WithError(err).Error("Failed to generate delete account OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate verification code",
		})
		return
	}

	otp := models.OTP{
		UserID:    &user.ID,
		Code:      otpCode,
		Type:      "delete_account",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := ac.db.Create(&otp).Error; err != nil {
		log.WithError(err).Error("Failed to save delete account OTP")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	if err := ac.emailService.SendOTPEmail(user.Email, otpCode, "delete_account"); err != nil {
		log.WithError(err).Error("Failed to send delete account email")

	}

	log.WithField("user_id", user.ID).Info("Delete account OTP sent")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Verification code sent to your email for account deletion",
	})
}

func (ac *AuthController) DeleteAccount(c *gin.Context) {
	log := logger.GetLogger()

	var req models.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind delete account request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		log.WithField("validation_errors", validationErrors).Warn("Delete account validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors,
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	var user models.User
	if err := ac.db.Where("id = ?", userID).First(&user).Error; err != nil {
		log.WithError(err).Error("User not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	var otp models.OTP
	if err := ac.db.Where("user_id = ? AND code = ? AND type = ? AND used = ? AND expires_at > ?",
		user.ID, req.OTPCode, "delete_account", false, time.Now()).First(&otp).Error; err != nil {
		log.WithError(err).Warn("Invalid or expired OTP for account deletion")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid or expired verification code",
		})
		return
	}

	tx := ac.db.Begin()

	if err := tx.Model(&otp).Update("used", true).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to mark delete account OTP as used")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to process request",
		})
		return
	}

	tokenString := c.GetHeader("Authorization")
	if tokenString != "" && len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]

		if expiresAt, err := ac.authService.ExtractTokenExpiry(tokenString); err == nil {
			blacklistEntry := models.TokenBlacklist{
				Token:     tokenString,
				ExpiresAt: expiresAt,
			}
			tx.Create(&blacklistEntry)
		}
	}

	if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&models.OTP{}).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to delete user OTPs")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete account",
		})
		return
	}

	if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&models.Profile{}).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to delete user profile")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete account",
		})
		return
	}

	if err := tx.Unscoped().Delete(&user).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to delete user account")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete account",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.WithError(err).Error("Failed to commit account deletion transaction")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete account",
		})
		return
	}

	log.WithField("user_id", user.ID).Info("User account successfully deleted")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Account successfully deleted",
	})
}
