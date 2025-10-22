# Bug Fix: Today's Target Returns Undefined (Timezone Issue)

## Problem Report

**Date**: October 23, 2025
**Reporter**: User
**Severity**: HIGH - Today's target not found despite data exists in database

### Symptoms

**Frontend Code:**
```javascript
// Get today's target
getToday: async () => {
  return fetchWithAuth(`${API_BASE_URL}/daily-targets/today`);
}

// Find today's target from list
const today = format(new Date(), "yyyy-MM-dd");
const targetToday = targetsRes.data.targets?.find((t: any) => {
  const targetDate = format(new Date(t.date), "yyyy-MM-dd");
  return targetDate === today;
});

console.log("Today's target:", targetToday); // undefined ❌
```

**Database State:**
```
Row 1: id=1, date=2025-10-19, income_target=1000000
Row 2: id=2, date=2025-10-20, income_target=2000000
Row 3: id=3, date=2025-10-22, income_target=1000000
Row 4: id=4, date=2025-10-23, income_target=5.00      ← TODAY!
```

**Result:**
- ❌ `GET /daily-targets/today` returns **500 Internal Server Error** or **404**
- ❌ Frontend gets **undefined** for today's target
- ✅ Data **exists** in database for 2025-10-23

---

## Root Cause Analysis

### Issue #1: Timezone Mismatch in Date Comparison

**Problem Code (Before):**
```go
// In GetTodayTarget()
today := time.Now().Truncate(24 * time.Hour)

// In GetDailyTargetByDate()
db.Where("user_id = ? AND date = ?", userID, date)
```

**Why it fails:**

1. **`time.Now().Truncate(24 * time.Hour)`** produces:
   ```
   2025-10-23 00:00:00 UTC
   ```

2. **Database stores with timezone** (depending on PostgreSQL settings):
   ```
   2025-10-23 00:00:00 +07:00  (Indonesia)
   2025-10-23 00:00:00 +08:00  (Singapore)
   ```

3. **Exact comparison fails:**
   ```sql
   WHERE date = '2025-10-23 00:00:00+00:00'  -- UTC
   -- Does NOT match:
   -- '2025-10-23 00:00:00+07:00'  -- Database value
   ```

4. **Result**: No rows found, returns 404

### Issue #2: Error Handling Returns 500 Instead of 404

**Problem Code (Before):**
```go
target, err := c.dailyTargetService.GetTodayTarget(userID)
if err != nil {
    // Always returns 500 for any error
    ctx.JSON(http.StatusInternalServerError, models.APIResponse{
        Success: false,
        Message: "Failed to retrieve today's target",
        Error:   err.Error(),
    })
    return
}
```

**Why it's wrong:**
- "Not Found" should return **404**, not **500**
- Frontend can't distinguish between "no data" vs "server error"
- User thinks system is broken when it's just missing data

### Issue #3: Frontend Uses Wrong Approach

**Problem Code:**
```javascript
// Gets ALL targets
const targetsRes = await api.finance.dailyTargets.getAll();

// Then filters manually
const targetToday = targetsRes.data.targets?.find((t: any) => {
  const targetDate = format(new Date(t.date), "yyyy-MM-dd");
  return targetDate === today;
});
```

**Why it's inefficient:**
1. Fetches **all** daily targets (could be 100+ records)
2. Transfers unnecessary data
3. Client-side filtering (slow for large datasets)
4. Doesn't use the dedicated `/today` endpoint

---

## Solution Implemented

### Fix #1: Use DATE() Function to Ignore Timezone

**File:** `internal/services/daily_target_service.go`

**GetTodayTarget() - BEFORE:**
```go
func (s *DailyTargetService) GetTodayTarget(userID uuid.UUID) (*models.DailyTarget, error) {
    today := time.Now().Truncate(24 * time.Hour)  // ❌ Timezone-sensitive

    target, err := s.GetDailyTargetByDate(userID, today)
    if err != nil {
        if err.Error() == "daily target not found" {
            return nil, fmt.Errorf("%w: no daily target found for today", utils.ErrNotFound)
        }
        return nil, err
    }

    // ...
}
```

**GetTodayTarget() - AFTER:**
```go
func (s *DailyTargetService) GetTodayTarget(userID uuid.UUID) (*models.DailyTarget, error) {
    // Use local time and extract date components
    now := time.Now()
    today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

    // Use DATE() SQL function to compare ONLY the date part
    var target models.DailyTarget
    if err := s.db.Where("user_id = ? AND DATE(date) = DATE(?)", userID, today).First(&target).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("%w: no daily target found for today", utils.ErrNotFound)
        }
        return nil, err
    }

    // Refresh actual values
    if err := s.updateActualValues(&target); err != nil {
        return nil, err
    }

    if err := s.db.Save(&target).Error; err != nil {
        return nil, err
    }

    return &target, nil
}
```

**GetDailyTargetByDate() - BEFORE:**
```go
func (s *DailyTargetService) GetDailyTargetByDate(userID uuid.UUID, date time.Time) (*models.DailyTarget, error) {
    var target models.DailyTarget
    // Exact timestamp match (fails with timezone differences)
    if err := s.db.Where("user_id = ? AND date = ?", userID, date).First(&target).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("daily target not found")
        }
        return nil, err
    }
    return &target, nil
}
```

**GetDailyTargetByDate() - AFTER:**
```go
func (s *DailyTargetService) GetDailyTargetByDate(userID uuid.UUID, date time.Time) (*models.DailyTarget, error) {
    var target models.DailyTarget
    // Use DATE() function to compare only date part, ignoring time and timezone
    if err := s.db.Where("user_id = ? AND DATE(date) = DATE(?)", userID, date).First(&target).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("%w: daily target not found", utils.ErrNotFound)
        }
        return nil, err
    }
    return &target, nil
}
```

**SQL Generated:**
```sql
-- BEFORE (fails with timezone mismatch)
SELECT * FROM daily_targets
WHERE user_id = ? AND date = '2025-10-23 00:00:00+00:00';

-- AFTER (works regardless of timezone)
SELECT * FROM daily_targets
WHERE user_id = ? AND DATE(date) = DATE('2025-10-23 00:00:00+07:00');
-- Extracts: 2025-10-23 = 2025-10-23 ✅
```

### Fix #2: Return 404 for "Not Found"

**File:** `internal/controllers/daily_target_controller.go`

**BEFORE:**
```go
target, err := c.dailyTargetService.GetTodayTarget(userID)
if err != nil {
    logger.GetLogger().Error("Failed to get today's target: ", err)
    ctx.JSON(http.StatusInternalServerError, models.APIResponse{  // ❌ Always 500
        Success: false,
        Message: "Failed to retrieve today's target",
        Error:   err.Error(),
    })
    return
}
```

**AFTER:**
```go
target, err := c.dailyTargetService.GetTodayTarget(userID)
if err != nil {
    // Return 404 if no target found for today
    if utils.IsNotFoundError(err) {
        ctx.JSON(http.StatusNotFound, models.APIResponse{  // ✅ Proper 404
            Success: false,
            Message: "No daily target found for today",
            Error:   err.Error(),
        })
        return
    }

    logger.GetLogger().Error("Failed to get today's target: ", err)
    ctx.JSON(http.StatusInternalServerError, models.APIResponse{
        Success: false,
        Message: "Failed to retrieve today's target",
        Error:   err.Error(),
    })
    return
}
```

---

## Recommended Frontend Changes

### Current Approach (Inefficient) ❌

```javascript
// Fetches ALL targets
const targetsRes = await api.finance.dailyTargets.getAll();

// Client-side filtering
const today = format(new Date(), "yyyy-MM-dd");
const targetToday = targetsRes.data.targets?.find((t: any) => {
  const targetDate = format(new Date(t.date), "yyyy-MM-dd");
  return targetDate === today;
});

console.log("Today's target:", targetToday);
```

**Problems:**
- 🐌 Fetches 100+ records when you only need 1
- 💾 Wastes bandwidth
- 🔍 Client-side search is slow
- ❌ Doesn't handle 404 properly

### Recommended Approach (Efficient) ✅

```javascript
// Use dedicated /today endpoint
try {
  const todayRes = await api.finance.dailyTargets.getToday();

  if (todayRes.success) {
    const targetToday = todayRes.data;
    console.log("Today's target:", targetToday);

    // Use the target
    setTodayTarget(targetToday);
  }
} catch (error) {
  if (error.response?.status === 404) {
    // No target for today - show "Create Target" button
    console.log("No target found for today, user needs to create one");
    setTodayTarget(null);
    setShowCreateButton(true);
  } else {
    // Real error - show error message
    console.error("Failed to fetch today's target:", error);
    setError("Failed to load today's target");
  }
}
```

**Benefits:**
- ⚡ Fast: Only fetches 1 record
- 💾 Efficient: Minimal data transfer
- 🎯 Server-side filtering (database indexed)
- ✅ Proper error handling

### Full Frontend Example

```javascript
import { format } from "date-fns";

// API service
const dailyTargetsAPI = {
  // Use this for today's target
  getToday: async () => {
    const response = await fetchWithAuth(`${API_BASE_URL}/daily-targets/today`);
    return response;
  },

  // Use this for listing all targets (paginated)
  getAll: async (params?: { page?: number; page_size?: number; date_from?: string; date_to?: string }) => {
    const queryParams = new URLSearchParams(params);
    const response = await fetchWithAuth(`${API_BASE_URL}/daily-targets?${queryParams}`);
    return response;
  },

  // Create new target
  create: async (data: CreateDailyTargetRequest) => {
    const response = await fetchWithAuth(`${API_BASE_URL}/daily-targets`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return response;
  }
};

// React Component
function TodayTargetSection() {
  const [todayTarget, setTodayTarget] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showCreateForm, setShowCreateForm] = useState(false);

  useEffect(() => {
    fetchTodayTarget();
  }, []);

  const fetchTodayTarget = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await dailyTargetsAPI.getToday();

      if (response.success) {
        setTodayTarget(response.data);
        setShowCreateForm(false);
      }
    } catch (err) {
      if (err.response?.status === 404) {
        // No target for today - this is OK
        console.log("No target found for today");
        setTodayTarget(null);
        setShowCreateForm(true); // Show create button
      } else {
        // Real error
        console.error("Error fetching today's target:", err);
        setError("Failed to load today's target");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCreateTarget = async (data) => {
    try {
      const response = await dailyTargetsAPI.create({
        date: new Date().toISOString().split('T')[0], // Today's date
        income_target: data.incomeTarget,
        expense_limit: data.expenseLimit,
        savings_target: data.savingsTarget,
        notes: data.notes || ""
      });

      if (response.success) {
        setTodayTarget(response.data);
        setShowCreateForm(false);
      }
    } catch (err) {
      console.error("Failed to create target:", err);
      alert("Failed to create daily target");
    }
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  if (!todayTarget && showCreateForm) {
    return (
      <div>
        <h2>No Target for Today</h2>
        <p>You haven't set a target for today yet.</p>
        <button onClick={() => setShowCreateForm(true)}>
          Create Today's Target
        </button>
        {/* Show create form here */}
      </div>
    );
  }

  return (
    <div>
      <h2>Today's Trading Target</h2>
      <div>
        <p>Income Target: Rp {todayTarget.income_target.toLocaleString()}</p>
        <p>Expense Limit: Rp {todayTarget.expense_limit.toLocaleString()}</p>
        <p>Actual Income: Rp {todayTarget.actual_income.toLocaleString()}</p>
        <p>Remaining: Rp {todayTarget.remaining_income.toLocaleString()}</p>
        <p>Win Rate: {todayTarget.win_rate.toFixed(2)}%</p>
      </div>
    </div>
  );
}
```

---

## Testing Instructions

### Test Case 1: Verify Today's Target Works

**Setup:**
```sql
-- Insert target for today (use your timezone)
INSERT INTO daily_targets (user_id, date, income_target, expense_limit, remaining_income, remaining_expense)
VALUES ('2201e461-2158-4677-a9bf-9f11eddb18eb', '2025-10-23', 1000000, 100000, 1000000, 100000);
```

**Test:**
```bash
# Restart server
go run cmd/main.go

# Test endpoint
curl -X GET http://localhost:8080/api/v1/finance/daily-targets/today \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected Response: 200 OK**
```json
{
  "success": true,
  "message": "Today's target retrieved successfully",
  "data": {
    "id": 4,
    "date": "2025-10-23",
    "income_target": 1000000,
    "expense_limit": 100000,
    "actual_income": 0,
    "remaining_income": 1000000,
    ...
  }
}
```

### Test Case 2: Verify 404 When No Target

**Setup:**
```sql
-- Delete today's target
DELETE FROM daily_targets WHERE DATE(date) = DATE('2025-10-23');
```

**Test:**
```bash
curl -X GET http://localhost:8080/api/v1/finance/daily-targets/today \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected Response: 404 Not Found**
```json
{
  "success": false,
  "message": "No daily target found for today",
  "error": "resource not found: no daily target found for today"
}
```

### Test Case 3: Verify Different Timezones

**Setup:**
```sql
-- Insert with explicit timezone
INSERT INTO daily_targets (user_id, date, income_target, expense_limit)
VALUES ('uuid', '2025-10-23 00:00:00+07:00', 1000000, 100000);
```

**Test:**
```bash
# Should find the target regardless of server timezone
curl -X GET http://localhost:8080/api/v1/finance/daily-targets/today \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected:** Returns 200 OK with target data

---

## Technical Details

### PostgreSQL DATE() Function

```sql
-- DATE() extracts only the date part, ignoring time and timezone
SELECT DATE('2025-10-23 00:00:00+00:00');  -- Result: 2025-10-23
SELECT DATE('2025-10-23 00:00:00+07:00');  -- Result: 2025-10-23
SELECT DATE('2025-10-23 15:30:45+08:00');  -- Result: 2025-10-23

-- Our query
WHERE DATE(date) = DATE('2025-10-23 00:00:00')
-- Matches all of these:
-- '2025-10-23 00:00:00+00:00' ✅
-- '2025-10-23 00:00:00+07:00' ✅
-- '2025-10-23 23:59:59+08:00' ✅
```

### Go time.Date() vs Truncate()

```go
// ❌ Truncate - keeps timezone, may not work
now := time.Now()  // 2025-10-23 15:30:00 +07:00
today := now.Truncate(24 * time.Hour)
// Result: 2025-10-23 00:00:00 +00:00 (UTC)

// ✅ time.Date - explicit date components
now := time.Now()  // 2025-10-23 15:30:00 +07:00
today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// Result: 2025-10-23 00:00:00 +07:00 (same timezone)
```

---

## Files Modified

1. **internal/services/daily_target_service.go**
   - `GetTodayTarget()`: Use `DATE()` function + proper error wrapping
   - `GetDailyTargetByDate()`: Use `DATE()` function for comparison

2. **internal/controllers/daily_target_controller.go**
   - `GetTodayTarget()`: Return 404 for NotFound errors

---

## Summary

**Problems Fixed:**
✅ Timezone mismatch in date comparison
✅ Error 500 instead of 404 for missing target
✅ Inefficient frontend data fetching

**Changes Made:**
✅ Use SQL `DATE()` function to ignore timezone
✅ Proper error handling (404 for NotFound)
✅ Documented efficient frontend pattern

**User Action Required:**
1. ✅ Restart Go backend server
2. ⏳ Update frontend to use `/today` endpoint directly
3. ⏳ Handle 404 error properly (show "Create Target" button)

**Status:** ✅ FIXED (backend) / ⏳ PENDING (frontend changes recommended)
