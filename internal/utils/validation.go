package utils

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// Validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("username", validateUsername)
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}

// validatePassword validates password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Minimum 8 characters
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Password must contain at least 3 of the 4 character types
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 3
}

// validateUsername validates username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Username should be 3-50 characters
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	// Username should only contain alphanumeric characters, underscore, and hyphen
	// Must start with a letter or number
	matched, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9_-]*$", username)
	return matched
}

// ValidateStruct validates a struct and returns formatted error messages
func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		errors[err.Field()] = getErrorMessage(err)
	}

	return errors
}

// getErrorMessage returns a user-friendly error message for validation errors
func getErrorMessage(err validator.FieldError) string {
	field := strings.ToLower(err.Field())

	switch err.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		return field + " must be at least " + err.Param() + " characters"
	case "max":
		return field + " must not exceed " + err.Param() + " characters"
	case "len":
		return field + " must be exactly " + err.Param() + " characters"
	case "eqfield":
		return field + " must match " + strings.ToLower(err.Param())
	case "password":
		return "Password must be at least 8 characters and contain at least 3 of the following: uppercase letter, lowercase letter, number, special character"
	case "username":
		return "Username must be 3-50 characters and can only contain letters, numbers, underscore, and hyphen. Must start with a letter or number"
	default:
		return field + " is invalid"
	}
}

// IsValidEmail validates email format using regex
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// SanitizeString removes potentially harmful characters from string
func SanitizeString(s string) string {
	// Remove leading and trailing whitespace
	s = strings.TrimSpace(s)

	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	return s
}

// NormalizeEmail normalizes email address (lowercase and trim)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeUsername normalizes username (lowercase and trim)
func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
