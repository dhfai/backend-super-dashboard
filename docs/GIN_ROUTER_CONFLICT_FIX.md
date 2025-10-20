# Gin Router Conflict Fix

## Problem

```
panic: ':targetId' in new path '/api/v1/finance/daily-targets/:targetId/trades'
conflicts with existing wildcard ':id' in existing prefix '/api/v1/finance/daily-targets/:id'
```

## Root Cause

Gin Router tidak bisa membedakan antara 2 wildcard parameters di level yang sama:
- `/daily-targets/:id` → untuk GetDailyTargetByID
- `/daily-targets/:targetId/trades` → untuk AddTradeToTarget

Ketika request datang ke `/daily-targets/123`, Gin tidak tahu apakah `123` adalah `:id` atau `:targetId`.

## Solution

### 1. Route Ordering Rule

Gin Router membutuhkan urutan route yang spesifik:
```
Specific Paths (e.g., /today, /stats)
    ↓
Wildcard with Nested Routes (e.g., /:id/trades)
    ↓
Simple Wildcard Routes (e.g., /:id)
```

### 2. Use Same Parameter Name

Gunakan parameter name yang sama (`:id`) untuk semua wildcard di level yang sama:

**Before (❌ ERROR):**
```go
targets.GET("/:id", dailyTargetController.GetDailyTargetByID)
targets.POST("/:targetId/trades", tradingActivityController.AddTradeToTarget)  // ❌ Conflict!
```

**After (✅ FIXED):**
```go
// Specific paths first
targets.GET("/today", dailyTargetController.GetTodayTarget)
targets.GET("/month-summary", dailyTargetController.GetCurrentMonthSummary)
targets.GET("/week-summary", dailyTargetController.GetWeekSummary)

// Wildcard with nested routes
targets.POST("/:id/trades", tradingActivityController.AddTradeToTarget)  // ✅ Same :id
targets.GET("/:id/trades", tradingActivityController.GetTradingActivitiesByTarget)
targets.POST("/:id/refresh", dailyTargetController.RefreshActualValues)

// Simple wildcard at the end
targets.GET("/:id", dailyTargetController.GetDailyTargetByID)
targets.PUT("/:id", dailyTargetController.UpdateDailyTarget)
targets.DELETE("/:id", dailyTargetController.DeleteDailyTarget)
```

### 3. Update Controllers

Controller harus menggunakan `ctx.Param("id")` bukan `ctx.Param("targetId")`:

**Before:**
```go
targetID, err := strconv.ParseUint(ctx.Param("targetId"), 10, 32)  // ❌
```

**After:**
```go
targetID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)  // ✅
```

## Files Modified

### 1. `internal/routes/routes.go`
- ✅ Reordered routes: specific → wildcard+nested → simple wildcard
- ✅ Changed `:targetId` to `:id`

### 2. `internal/controllers/trading_activity_controller.go`
- ✅ Changed `ctx.Param("targetId")` to `ctx.Param("id")` in `AddTradeToTarget()`
- ✅ Changed `ctx.Param("targetId")` to `ctx.Param("id")` in `GetTradingActivitiesByTarget()`

## Final Route Structure

```
POST   /api/v1/finance/daily-targets
GET    /api/v1/finance/daily-targets
GET    /api/v1/finance/daily-targets/today
GET    /api/v1/finance/daily-targets/month-summary
GET    /api/v1/finance/daily-targets/week-summary
POST   /api/v1/finance/daily-targets/:id/trades          ✅
GET    /api/v1/finance/daily-targets/:id/trades          ✅
POST   /api/v1/finance/daily-targets/:id/refresh
GET    /api/v1/finance/daily-targets/:id
PUT    /api/v1/finance/daily-targets/:id
DELETE /api/v1/finance/daily-targets/:id
```

## Testing

```bash
# Start server
go run cmd/main.go

# Should start without panic
# INFO [2025-10-19] Server starting on :8080
```

## Usage Examples

All endpoints still work the same way, just using consistent parameter names:

```bash
# Add trade to daily target ID 1
POST /api/v1/finance/daily-targets/1/trades

# Get trades for daily target ID 1
GET /api/v1/finance/daily-targets/1/trades

# Get daily target by ID
GET /api/v1/finance/daily-targets/1

# Refresh daily target ID 1
POST /api/v1/finance/daily-targets/1/refresh
```

## Lesson Learned

✅ **Always use the same parameter name** for wildcards at the same route level
✅ **Order routes correctly**: Specific → Wildcard+Nested → Simple Wildcard
✅ **Test route registration** during development to catch conflicts early

---

**Status:** ✅ FIXED - Server should now start without panic
