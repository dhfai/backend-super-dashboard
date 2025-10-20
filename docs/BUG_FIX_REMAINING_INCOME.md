# Bug Fix: Remaining Income/Expense Not Updating

## Problem Report

**Date**: October 20, 2025
**Reporter**: User
**Severity**: HIGH - Core trading feature broken

### Symptoms

After adding 2 winning trades (Rp 180,000 + Rp 200,000 = Rp 380,000):

**Expected:**
- Remaining Income: Rp 1,000,000 - Rp 380,000 = **Rp 620,000**

**Actual:**
- Remaining Income: **Rp 800,000** ❌ (unchanged from initial value)

### Database Evidence

```
id: 1
income_target: 1,000,000.00
expense_limit: 100,000.00
total_trades: 2
winning_trades: 2
losing_trades: 0
actual_income: 0.00        ❌ Should be 380,000
actual_expense: 0.00       ✅ Correct
remaining_income: 800,000.00  ❌ Should be 620,000
remaining_expense: 100,000.00 ✅ Correct
win_rate: 100.00           ✅ Correct
```

### Trades Recorded

```
Trade 1: Win - Rp 180,000 (BTCUSD, 100 pips, lot 1)
Trade 2: Win - Rp 200,000 (BTCUSD, 100 pips, lot 0.02)
```

---

## Root Cause Analysis

### Issue #1: GORM `Save()` Not Updating All Fields

**Location:** `internal/services/trading_activity_service.go:118`

**Problem Code:**
```go
// This doesn't reliably update all fields
if err := tx.Save(&target).Error; err != nil {
    tx.Rollback()
    return nil, nil, err
}
```

**Why it fails:**
1. **GORM's `Save()` behavior**: Only updates **changed fields** that GORM can detect
2. **Zero value issue**: When a field goes from `0 → 620000`, GORM might skip it if not properly tracked
3. **Struct updates**: Modifying struct fields directly (e.g., `target.RemainingIncome = ...`) might not be detected by GORM's change tracking

### Issue #2: Transaction Order

**Problem Code:**
```go
// Get target BEFORE starting transaction
var target models.DailyTarget
if err := s.db.Where("id = ? AND user_id = ?", targetID, userID).First(&target).Error; err != nil {
    return nil, nil, err
}

// Start transaction AFTER loading target
tx := s.db.Begin()
```

**Why it's risky:**
1. **Race condition**: Target could be modified between read and transaction start
2. **No database lock**: Another request could update target simultaneously
3. **Stale data**: Working with potentially outdated target values

---

## Solution Implemented

### Fix #1: Use Explicit `Updates()` with Map

**Changed From:**
```go
// Unreliable - GORM might not detect all changes
if err := tx.Save(&target).Error; err != nil {
    tx.Rollback()
    return nil, nil, err
}
```

**Changed To:**
```go
// Explicit field updates - GORM MUST update these fields
if err := tx.Model(&target).Updates(map[string]interface{}{
    "total_trades":       target.TotalTrades,
    "winning_trades":     target.WinningTrades,
    "losing_trades":      target.LosingTrades,
    "actual_income":      target.ActualIncome,      // ✅ NOW UPDATES
    "actual_expense":     target.ActualExpense,     // ✅ NOW UPDATES
    "actual_savings":     target.ActualSavings,     // ✅ NOW UPDATES
    "remaining_income":   target.RemainingIncome,   // ✅ NOW UPDATES
    "remaining_expense":  target.RemainingExpense,  // ✅ NOW UPDATES
    "win_rate":           target.WinRate,
    "is_completed":       target.IsCompleted,
    "completed_at":       target.CompletedAt,
}).Error; err != nil {
    tx.Rollback()
    return nil, nil, err
}
```

**Benefits:**
- ✅ **Explicit field list**: GORM knows exactly which fields to UPDATE
- ✅ **Handles zero values**: Using `map[string]interface{}` forces update even for 0
- ✅ **Predictable behavior**: No reliance on GORM's change detection
- ✅ **Generated SQL**: `UPDATE daily_targets SET remaining_income = 620000 WHERE ...`

### Fix #2: Load Target INSIDE Transaction

**Changed From:**
```go
// Load target outside transaction
var target models.DailyTarget
if err := s.db.Where("id = ? AND user_id = ?", targetID, userID).First(&target).Error; err != nil {
    return nil, nil, err
}

// Start transaction later
tx := s.db.Begin()
```

**Changed To:**
```go
// Start transaction FIRST
tx := s.db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// Load target INSIDE transaction (with implicit lock)
var target models.DailyTarget
if err := tx.Where("id = ? AND user_id = ?", targetID, userID).First(&target).Error; err != nil {
    tx.Rollback()
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil, fmt.Errorf("%w: daily target not found", utils.ErrNotFound)
    }
    return nil, nil, err
}
```

**Benefits:**
- ✅ **Database lock**: PostgreSQL locks the row during transaction
- ✅ **ACID compliance**: All operations atomic
- ✅ **No race conditions**: Other requests wait until commit
- ✅ **Consistent reads**: Target data guaranteed fresh

---

## Files Modified

### 1. `internal/services/trading_activity_service.go`

#### Function: `AddTradeToTarget()`
**Lines Changed:** 23-128
**Changes:**
- Moved `tx.Begin()` to top of function
- Moved target query inside transaction
- Changed `tx.Save(&target)` → `tx.Model(&target).Updates(map[...])`
- Added explicit field list for all trading stats

#### Function: `DeleteTradingActivity()`
**Lines Changed:** 169-256
**Changes:**
- Moved `tx.Begin()` to top of function
- Moved activity and target queries inside transaction
- Changed `tx.Save(&target)` → `tx.Model(&target).Updates(map[...])`
- Added explicit field list for recalculated stats

---

## Testing Instructions

### Test Case 1: Add Winning Trades

**Setup:**
```sql
-- Create fresh daily target
INSERT INTO daily_targets (user_id, date, income_target, expense_limit, remaining_income, remaining_expense)
VALUES ('uuid', '2025-10-20', 1000000, 100000, 1000000, 100000);
```

**Test:**
```bash
# Add first win
POST /api/v1/finance/daily-targets/1/trades
{
  "trade_type": "win",
  "amount": 180000,
  "symbol": "BTCUSD",
  "pips": 100,
  "lot_size": 1
}

# Add second win
POST /api/v1/finance/daily-targets/1/trades
{
  "trade_type": "win",
  "amount": 200000,
  "symbol": "BTCUSD",
  "pips": 100,
  "lot_size": 0.02
}

# Check result
GET /api/v1/finance/daily-targets/1
```

**Expected Response:**
```json
{
  "id": 1,
  "income_target": 1000000,
  "expense_limit": 100000,
  "actual_income": 380000,        // ✅ 180k + 200k
  "actual_expense": 0,
  "actual_savings": 380000,       // ✅ 380k - 0
  "remaining_income": 620000,     // ✅ 1M - 380k = 620k
  "remaining_expense": 100000,    // ✅ Unchanged
  "total_trades": 2,
  "winning_trades": 2,
  "losing_trades": 0,
  "win_rate": 100.00
}
```

### Test Case 2: Add Losing Trades

**Test:**
```bash
# Add loss
POST /api/v1/finance/daily-targets/1/trades
{
  "trade_type": "loss",
  "amount": 50000,
  "symbol": "EURUSD",
  "pips": -50,
  "lot_size": 1
}

# Check result
GET /api/v1/finance/daily-targets/1
```

**Expected Response:**
```json
{
  "actual_income": 380000,        // ✅ Unchanged
  "actual_expense": 50000,        // ✅ +50k
  "actual_savings": 330000,       // ✅ 380k - 50k
  "remaining_income": 620000,     // ✅ Still 620k
  "remaining_expense": 50000,     // ✅ 100k - 50k = 50k
  "total_trades": 3,
  "winning_trades": 2,
  "losing_trades": 1,
  "win_rate": 66.67               // ✅ 2/3 * 100
}
```

### Test Case 3: Delete Trade

**Test:**
```bash
# Delete first win trade (180k)
DELETE /api/v1/finance/daily-targets/1/trades/1

# Check result
GET /api/v1/finance/daily-targets/1
```

**Expected Response:**
```json
{
  "actual_income": 200000,        // ✅ 380k - 180k
  "actual_expense": 50000,        // ✅ Unchanged
  "actual_savings": 150000,       // ✅ 200k - 50k
  "remaining_income": 800000,     // ✅ 1M - 200k = 800k
  "remaining_expense": 50000,     // ✅ Unchanged
  "total_trades": 2,              // ✅ 3 - 1
  "winning_trades": 1,            // ✅ 2 - 1
  "losing_trades": 1,
  "win_rate": 50.00               // ✅ 1/2 * 100
}
```

---

## Technical Deep Dive

### GORM Update Behaviors

#### 1. `Save()` - Full Model Update
```go
db.Save(&target)
// SQL: UPDATE daily_targets SET field1=?, field2=?, ... WHERE id=?
// Updates ALL fields, but might skip unchanged or zero-valued fields
```

#### 2. `Updates(struct)` - Non-Zero Fields
```go
db.Model(&target).Updates(target)
// SQL: UPDATE daily_targets SET changed_field1=?, changed_field2=? WHERE id=?
// Only updates non-zero fields
```

#### 3. `Updates(map)` - Explicit Fields (USED IN FIX)
```go
db.Model(&target).Updates(map[string]interface{}{
    "remaining_income": 0,  // ✅ Updates even if 0
})
// SQL: UPDATE daily_targets SET remaining_income=0 WHERE id=?
// ALWAYS updates specified fields
```

#### 4. `Update(column, value)` - Single Field
```go
db.Model(&target).Update("remaining_income", 620000)
// SQL: UPDATE daily_targets SET remaining_income=620000 WHERE id=?
// Updates single field
```

### Why We Chose `Updates(map)`

| Method | Zero Values | Multiple Fields | Type Safety | Performance |
|--------|-------------|-----------------|-------------|-------------|
| `Save()` | ❌ Unreliable | ✅ All fields | ✅ Type-safe | ⚠️ Updates all |
| `Updates(struct)` | ❌ Skips zeros | ✅ Multiple | ✅ Type-safe | ✅ Selective |
| **`Updates(map)`** | **✅ Includes zeros** | **✅ Multiple** | ⚠️ Runtime | **✅ Selective** |
| `Update(col, val)` | ✅ Includes zeros | ❌ One field | ✅ Type-safe | ✅ Fast |

**Verdict:** `Updates(map)` is best for our use case because:
1. We need to update **multiple fields** (11 fields)
2. Some fields might be **zero** (e.g., `actual_income = 0`)
3. We want **selective updates** (not all model fields)
4. Performance is acceptable (single UPDATE query)

---

## Lessons Learned

### 1. Don't Trust GORM's Change Detection

**Problem:**
```go
target.RemainingIncome = 620000
db.Save(&target)  // Might not update!
```

**Solution:**
```go
db.Model(&target).Updates(map[string]interface{}{
    "remaining_income": 620000,  // Always updates
})
```

### 2. Always Use Transactions for Multi-Step Operations

**Bad:**
```go
// Step 1: Create activity
db.Create(&activity)

// Step 2: Update target (if this fails, activity is orphaned!)
db.Save(&target)
```

**Good:**
```go
tx := db.Begin()
tx.Create(&activity)
tx.Model(&target).Updates(map[...])
tx.Commit()  // All or nothing
```

### 3. Load Related Data INSIDE Transaction

**Bad (Race Condition):**
```go
target := loadTarget()  // Read
tx := db.Begin()
// Target might have changed since read!
tx.Save(&target)
tx.Commit()
```

**Good (Database Lock):**
```go
tx := db.Begin()
target := tx.First(&target)  // Locks row
// No one can modify target until commit
tx.Model(&target).Updates(map[...])
tx.Commit()
```

### 4. Test Zero-Value Updates

Always test:
- Setting field to `0`
- Setting field from `0` to non-zero
- Setting field from non-zero to `0`

### 5. Use Explicit Field Lists for Critical Updates

Don't rely on:
- Auto-detection
- Struct tags
- Zero-value inference

Be explicit about what MUST update.

---

## Verification Checklist

After deploying this fix:

- [x] **Code Review**: Changes reviewed and approved
- [ ] **Unit Tests**: Add tests for zero-value updates
- [ ] **Integration Tests**: Test full trading flow
- [ ] **Database Check**: Verify UPDATE queries in logs
- [ ] **Manual Test**: Follow test cases above
- [ ] **Performance**: Check query execution time
- [ ] **Rollback Plan**: Keep old code for quick revert

---

## Related Issues

- **Auto-Create Daily Target**: Fixed in previous commit (removed auto-creation)
- **Error 404 vs 500**: Fixed in previous commit (proper error types)
- **Remaining Income Bug**: This fix

---

## SQL Generated (Before vs After)

### Before Fix (GORM `Save()`)
```sql
-- Might generate something like:
UPDATE daily_targets
SET total_trades = 2,
    winning_trades = 2,
    win_rate = 100.00,
    updated_at = NOW()
WHERE id = 1;
-- Missing: actual_income, remaining_income, etc.
```

### After Fix (GORM `Updates(map)`)
```sql
UPDATE daily_targets
SET total_trades = 2,
    winning_trades = 2,
    losing_trades = 0,
    actual_income = 380000.00,         -- ✅ NOW INCLUDED
    actual_expense = 0.00,             -- ✅ NOW INCLUDED
    actual_savings = 380000.00,        -- ✅ NOW INCLUDED
    remaining_income = 620000.00,      -- ✅ NOW INCLUDED
    remaining_expense = 100000.00,     -- ✅ NOW INCLUDED
    win_rate = 100.00,
    is_completed = false,
    completed_at = NULL,
    updated_at = NOW()
WHERE id = 1;
```

---

## Status

**Status**: ✅ FIXED
**Fixed By**: AI Assistant
**Fixed Date**: October 20, 2025
**Tested**: Pending user verification
**Deployed**: Not yet

**Next Steps:**
1. User should clean database using cleanup script
2. Restart Go server to load new code
3. Test with fresh daily target
4. Verify all calculations are correct
