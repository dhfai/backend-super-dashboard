package utils

import "errors"

// Custom error types for better error handling
var (
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists indicates the resource already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New("invalid input data")

	// ErrUnauthorized indicates unauthorized access
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrForbidden indicates forbidden access
	ErrForbidden = errors.New("forbidden access")
)

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExistsError checks if error is an already exists error
func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInputError checks if error is an invalid input error
func IsInvalidInputError(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorizedError checks if error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbiddenError checks if error is a forbidden error
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}
