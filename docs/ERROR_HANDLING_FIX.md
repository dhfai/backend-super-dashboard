# Error Handling Fix - Daily Target Already Exists

## Problem

When trying to create a daily target for a date that already has a target, the API returned:
- **HTTP Status**: 500 Internal Server Error ❌
- **Expected Status**: 409 Conflict ✅

### Error Log:
```log
2025/10/19 20:56:28 ERROR [2025-10-19 20:56:28] Failed to create daily target: daily target already exists for this date
ERROR [2025-10-19 20:56:28] HTTP Request status=500
```

## Root Cause

The service layer was returning a generic error:
```go
return nil, errors.New("daily target already exists for this date")
```

The controller was treating ALL errors as Internal Server Error (500):
```go
if err != nil {
    ctx.JSON(http.StatusInternalServerError, ...)  // Always 500!
}
```

## Solution

### 1. Created Custom Error Types (`internal/utils/errors.go`)

```go
var (
    ErrNotFound = errors.New("resource not found")
    ErrAlreadyExists = errors.New("resource already exists")
    ErrInvalidInput = errors.New("invalid input data")
    ErrUnauthorized = errors.New("unauthorized access")
    ErrForbidden = errors.New("forbidden access")
)

// Helper functions
func IsAlreadyExistsError(err error) bool {
    return errors.Is(err, ErrAlreadyExists)
}
```

### 2. Updated Service to Wrap Errors (`daily_target_service.go`)

```go
if err := s.db.Where("user_id = ? AND date = ?", userID, req.Date).First(&existing).Error; err == nil {
    return nil, fmt.Errorf("%w: daily target already exists for this date", utils.ErrAlreadyExists)
}
```

### 3. Updated Controller to Handle Error Types (`daily_target_controller.go`)

```go
if err != nil {
    // Handle specific error types
    if utils.IsAlreadyExistsError(err) {
        ctx.JSON(http.StatusConflict, models.APIResponse{  // 409 Conflict
            Success: false,
            Message: "Daily target already exists for this date",
            Error:   err.Error(),
        })
        return
    }

    // Generic internal error
    ctx.JSON(http.StatusInternalServerError, ...)  // 500 for unexpected errors
}
```

## HTTP Status Code Mapping

| Error Type | HTTP Status | Use Case |
|------------|-------------|----------|
| `ErrAlreadyExists` | 409 Conflict | Resource already exists |
| `ErrNotFound` | 404 Not Found | Resource not found |
| `ErrInvalidInput` | 400 Bad Request | Invalid input data |
| `ErrUnauthorized` | 401 Unauthorized | Missing/invalid authentication |
| `ErrForbidden` | 403 Forbidden | Insufficient permissions |
| Other errors | 500 Internal Server Error | Unexpected errors |

## Testing

### Before Fix:
```bash
POST /api/v1/finance/daily-targets
{
  "date": "2025-10-19T00:00:00Z",
  "income_target": 500.00
}

# Response: 500 Internal Server Error ❌
```

### After Fix:
```bash
POST /api/v1/finance/daily-targets
{
  "date": "2025-10-19T00:00:00Z",
  "income_target": 500.00
}

# First time: 201 Created ✅
# Second time: 409 Conflict ✅
{
  "success": false,
  "message": "Daily target already exists for this date",
  "error": "resource already exists: daily target already exists for this date"
}
```

## Benefits

1. ✅ **Proper HTTP Status Codes**: Clients can distinguish between different error types
2. ✅ **Better Error Handling**: Error types can be checked programmatically
3. ✅ **Consistent Error Messages**: All services can use the same error types
4. ✅ **RESTful Compliance**: Follows HTTP status code best practices
5. ✅ **Better Client Experience**: Frontend can show appropriate messages based on status code

## Next Steps (Optional Improvements)

1. Apply the same pattern to other services:
   - `transaction_service.go`
   - `financial_goal_service.go`
   - `backtest_strategy_service.go`
   - `budget_service.go`

2. Create a middleware to automatically handle error types globally

3. Add structured error responses with error codes:
```go
type ErrorResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Error   string `json:"error"`
    Code    string `json:"code"` // e.g., "ALREADY_EXISTS"
}
```

## References

- **HTTP Status Codes**: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
- **Go Error Wrapping**: https://go.dev/blog/go1.13-errors
- **RESTful API Design**: https://restfulapi.net/http-status-codes/
