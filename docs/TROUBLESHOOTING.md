# Troubleshooting Guide

## Issue 1: Error 500 untuk "Not Found" Resources

### Problem
```
ERROR [2025-10-19 21:39:30] Failed to update daily target: resource not found: daily target not found
ERROR [2025-10-19 21:39:30] HTTP Request status=500
```

### Root Cause
Controller tidak handle `ErrNotFound` error dengan benar, semua error dikembalikan sebagai **500 Internal Server Error** padahal seharusnya **404 Not Found**.

### Solution Applied

Updated `internal/controllers/daily_target_controller.go` untuk 3 methods:

#### 1. UpdateDailyTarget
```go
target, err := c.dailyTargetService.UpdateDailyTarget(userID, uint(targetID), &req)
if err != nil {
    // Handle specific error types
    if utils.IsNotFoundError(err) {
        ctx.JSON(http.StatusNotFound, models.APIResponse{  // 404 ✅
            Success: false,
            Message: "Daily target not found",
            Error:   err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusInternalServerError, ...)  // 500 for other errors
}
```

#### 2. DeleteDailyTarget
```go
if err := c.dailyTargetService.DeleteDailyTarget(userID, uint(targetID)); err != nil {
    if utils.IsNotFoundError(err) {
        ctx.JSON(http.StatusNotFound, ...)  // 404 ✅
        return
    }
    ctx.JSON(http.StatusInternalServerError, ...)
}
```

#### 3. RefreshActualValues
```go
target, err := c.dailyTargetService.RefreshActualValues(userID, uint(targetID))
if err != nil {
    if utils.IsNotFoundError(err) {
        ctx.JSON(http.StatusNotFound, ...)  // 404 ✅
        return
    }
    ctx.JSON(http.StatusInternalServerError, ...)
}
```

### Testing

**Before (❌):**
```bash
PUT /api/v1/finance/daily-targets/999
# Response: 500 Internal Server Error
```

**After (✅):**
```bash
PUT /api/v1/finance/daily-targets/999
# Response: 404 Not Found
{
  "success": false,
  "message": "Daily target not found",
  "error": "resource not found: daily target not found"
}
```

---

## Issue 2: Data Muncul Otomatis di Database

### Problem
User belum create data tapi sudah ada row di tabel `daily_targets` (ID 7).

### Possible Causes

1. **Previous Test Data**
   - Data dari testing API sebelumnya
   - POST request yang sudah dikirim sebelumnya

2. **Browser/Frontend Auto-Request**
   - Frontend melakukan request otomatis
   - Browser refresh dengan previous request

3. **Duplicate Request**
   - Double click pada submit button
   - Multiple POST requests

### How to Check

#### 1. Check Database Directly
```sql
SELECT * FROM daily_targets
ORDER BY created_at DESC;
```

#### 2. Check Application Logs
```bash
# Cari POST request ke daily-targets
grep "POST.*daily-targets" logs.txt
```

#### 3. Check Browser Network Tab
- Open DevTools → Network
- Filter: `/daily-targets`
- Check semua POST requests

### Solutions

#### Option 1: Delete Test Data via SQL
```sql
-- Delete all financial test data
DELETE FROM trading_activities;
DELETE FROM daily_targets;
DELETE FROM transactions;
DELETE FROM financial_goals;
DELETE FROM backtest_strategies;
DELETE FROM budgets;

-- Verify
SELECT COUNT(*) FROM daily_targets;
```

#### Option 2: Use Cleanup Script
```bash
# Make script executable
chmod +x scripts/cleanup-financial-data.sh

# Set database credentials
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=filestore

# Run cleanup
./scripts/cleanup-financial-data.sh
```

#### Option 3: Delete via API
```bash
# Get all daily targets
GET /api/v1/finance/daily-targets

# Delete each one
DELETE /api/v1/finance/daily-targets/7
```

### Prevention

#### 1. Add Confirmation in Frontend
```javascript
// Before creating daily target
if (confirm('Create daily target?')) {
  createDailyTarget(data);
}
```

#### 2. Disable Submit Button After Click
```javascript
const handleSubmit = async () => {
  setLoading(true);
  try {
    await createDailyTarget(data);
  } finally {
    setLoading(false);
  }
}
```

#### 3. Check for Existing Target
```javascript
// Before creating, check if target exists for date
const existingTarget = await getDailyTarget(date);
if (existingTarget) {
  alert('Target already exists for this date!');
  return;
}
```

### Debugging Steps

1. **Check Created Date**
   ```sql
   SELECT id, date, created_at
   FROM daily_targets
   WHERE id = 7;
   ```

2. **Check Application Logs**
   ```bash
   # Find when ID 7 was created
   grep "Daily target created successfully" logs.txt | grep "id.*7"
   ```

3. **Check User ID**
   ```sql
   SELECT id, user_id, date
   FROM daily_targets;
   ```
   Verify it's your user_id.

---

## Common HTTP Status Codes Reference

| Status | When to Use |
|--------|-------------|
| **200 OK** | Successful GET, PUT, PATCH |
| **201 Created** | Successful POST (resource created) |
| **204 No Content** | Successful DELETE |
| **400 Bad Request** | Invalid input, validation error |
| **401 Unauthorized** | Missing/invalid authentication |
| **403 Forbidden** | Authenticated but no permission |
| **404 Not Found** | Resource doesn't exist ✅ |
| **409 Conflict** | Resource already exists ✅ |
| **422 Unprocessable Entity** | Validation error (alternative to 400) |
| **500 Internal Server Error** | Unexpected server error only |

---

## Summary

### Fixed:
✅ UpdateDailyTarget returns **404** when target not found
✅ DeleteDailyTarget returns **404** when target not found
✅ RefreshActualValues returns **404** when target not found
✅ Created cleanup script for test data

### To Clean Database:
```bash
# Option 1: SQL
DELETE FROM trading_activities;
DELETE FROM daily_targets;

# Option 2: Script
./scripts/cleanup-financial-data.sh

# Option 3: API
DELETE /api/v1/finance/daily-targets/7
```

### Prevention:
- Add frontend confirmation
- Disable button after click
- Check existing data before create
- Monitor application logs
