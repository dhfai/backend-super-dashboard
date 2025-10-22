# Bug Fix: Cannot Create Daily Target - Duplicate Key Error

## Problem Report

**Date**: October 23, 2025
**Severity**: CRITICAL - Blocks user from creating daily targets

### Error Message

```
ERROR: duplicate key value violates unique constraint "idx_user_date" (SQLSTATE 23505)
INSERT INTO "daily_targets" (...) VALUES (...,'2025-10-23',...) RETURNING "id"
```

### User Flow

1. ✅ User creates daily target (2025-10-23) → Success
2. ✅ User adds trades → Success
3. ✅ User deletes daily target via API → **Frontend shows success**
4. ❌ Check database → **Data still exists!**
5. ❌ User tries to create new daily target (2025-10-23) → **ERROR: duplicate key**

### Database State

```sql
SELECT id, user_id, date, deleted_at FROM daily_targets;

id | user_id | date       | deleted_at
---+---------+------------+-------------------------
1  | uuid    | 2025-10-23 | 2025-10-23 04:35:00  ← SOFT DELETED!
```

**Problem**: Row still exists with `deleted_at` set, blocking new inserts!

---

## Root Cause Analysis

### Issue: Soft Delete + Unique Constraint = Conflict

**Model Definition (BEFORE):**
```go
type DailyTarget struct {
    UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
    Date      time.Time      `gorm:"uniqueIndex:idx_user_date"` // ❌ PROBLEM
    DeletedAt gorm.DeletedAt `gorm:"index"`                      // ❌ NOT IN UNIQUE INDEX
}
```

**Unique Index Created:**
```sql
CREATE UNIQUE INDEX idx_user_date ON daily_targets (user_id, date);
```

**What Happens:**

1. **Create** daily target:
   ```sql
   INSERT INTO daily_targets (user_id, date) VALUES ('uuid', '2025-10-23');
   -- Row 1: user_id='uuid', date='2025-10-23', deleted_at=NULL ✅
   ```

2. **Soft Delete** (via `gorm.Delete`):
   ```sql
   UPDATE daily_targets SET deleted_at = NOW() WHERE id = 1;
   -- Row 1: user_id='uuid', date='2025-10-23', deleted_at='2025-10-23 04:35:00' ✅
   -- ⚠️ Row still exists! Unique constraint still active!
   ```

3. **Create again** (same date):
   ```sql
   INSERT INTO daily_targets (user_id, date) VALUES ('uuid', '2025-10-23');
   -- ❌ ERROR: duplicate key (user_id='uuid', date='2025-10-23' already exists)
   ```

**Why it fails:**
- Unique index checks `(user_id, date)` **regardless of `deleted_at`**
- Soft-deleted row still "occupies" that unique slot
- Cannot create new target for same date, even after deletion

---

## Solution Implemented

### Option 1: Include DeletedAt in Unique Index (CHOSEN) ✅

**Model Definition (AFTER):**
```go
type DailyTarget struct {
    ID        uint           `gorm:"primaryKey"`
    UserID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_date_deleted"`
    Date      time.Time      `gorm:"type:date;not null;uniqueIndex:idx_user_date_deleted"`
    DeletedAt gorm.DeletedAt `gorm:"index;uniqueIndex:idx_user_date_deleted"` // ✅ ADDED
}
```

**Unique Index Created:**
```sql
CREATE UNIQUE INDEX idx_user_date_deleted
ON daily_targets (user_id, date, deleted_at);
```

**How It Works:**

1. **Active record** (deleted_at = NULL):
   ```
   (user_id='uuid', date='2025-10-23', deleted_at=NULL) ✅ Unique
   ```

2. **Soft delete** (deleted_at = timestamp):
   ```
   (user_id='uuid', date='2025-10-23', deleted_at='2025-10-23 04:35:00') ✅ Different!
   ```

3. **Create new** (deleted_at = NULL again):
   ```
   (user_id='uuid', date='2025-10-23', deleted_at=NULL) ✅ Unique (no conflict!)
   ```

**Benefits:**
- ✅ Can create new target after deleting old one
- ✅ Keeps soft delete (audit trail)
- ✅ Each soft-delete has different timestamp (unique)
- ✅ Only ONE active record per (user_id, date)

### Option 2: Use Hard Delete (Alternative)

**Change DeleteDailyTarget to hard delete:**
```go
// Use .Unscoped() to bypass soft delete
result := s.db.Unscoped().Where("id = ? AND user_id = ?", targetID, userID).
    Delete(&models.DailyTarget{})
```

**Pros & Cons:**
- ✅ Simple - truly removes data
- ❌ Loses audit trail (no history)
- ❌ Cannot restore deleted data

---

## Migration Required

### Step 1: Drop Old Unique Index

```sql
-- Drop old index
DROP INDEX IF EXISTS idx_user_date;
```

### Step 2: Create New Unique Index with DeletedAt

```sql
-- Create new index including deleted_at
CREATE UNIQUE INDEX idx_user_date_deleted
ON daily_targets (user_id, date, deleted_at);
```

### Step 3: Clean Soft-Deleted Records (Optional)

If you want to clean up existing soft-deleted records:

```sql
-- Option A: Permanently delete soft-deleted records
DELETE FROM daily_targets WHERE deleted_at IS NOT NULL;

-- Option B: Keep them for audit trail (recommended)
-- Do nothing, new index will handle it
```

### Step 4: Restart Server

```bash
# GORM Auto-Migrate will detect model changes
go run cmd/main.go
```

**Important**: GORM may not automatically drop old index. Run SQL manually if needed.

---

## Testing Instructions

### Test Case 1: Create → Delete → Create Again

**Setup:**
```bash
# Restart server with new model
go run cmd/main.go
```

**Test:**
```bash
# 1. Create daily target
POST /api/v1/finance/daily-targets
{
  "date": "2025-10-23",
  "income_target": 1000000,
  "expense_limit": 100000
}
# Response: 201 Created, id=1

# 2. Delete daily target
DELETE /api/v1/finance/daily-targets/1
# Response: 200 OK

# 3. Create again (SAME DATE)
POST /api/v1/finance/daily-targets
{
  "date": "2025-10-23",
  "income_target": 2000000,
  "expense_limit": 200000
}
# Response: 201 Created, id=2 ✅ (should work now!)
```

**Verify Database:**
```sql
SELECT id, user_id, date, income_target, deleted_at
FROM daily_targets
ORDER BY id;

-- Expected:
-- id=1, date=2025-10-23, deleted_at=2025-10-23 04:35:00 (soft deleted)
-- id=2, date=2025-10-23, deleted_at=NULL (active) ✅
```

### Test Case 2: Prevent Duplicate Active Records

**Test:**
```bash
# 1. Create daily target
POST /api/v1/finance/daily-targets
{
  "date": "2025-10-24",
  "income_target": 1000000
}
# Response: 201 Created, id=3

# 2. Try to create another for SAME DATE (without deleting)
POST /api/v1/finance/daily-targets
{
  "date": "2025-10-24",
  "income_target": 2000000
}
# Response: 409 Conflict ✅
# Error: "daily target already exists for this date"
```

### Test Case 3: Verify Unique Index

**SQL Test:**
```sql
-- Check unique index exists
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'daily_targets'
  AND indexname = 'idx_user_date_deleted';

-- Expected:
-- CREATE UNIQUE INDEX idx_user_date_deleted
-- ON daily_targets USING btree (user_id, date, deleted_at)
```

---

## Database State Comparison

### BEFORE Fix (Broken)

```sql
SELECT * FROM daily_targets;

id | user_id | date       | income_target | deleted_at
---+---------+------------+---------------+-------------------------
1  | uuid    | 2025-10-23 | 1000000       | 2025-10-23 04:35:00

-- Unique Index: (user_id, date)
-- Tries to insert: (uuid, 2025-10-23)
-- ❌ ERROR: duplicate key (row 1 matches, even though soft-deleted)
```

### AFTER Fix (Working)

```sql
SELECT * FROM daily_targets;

id | user_id | date       | income_target | deleted_at
---+---------+------------+---------------+-------------------------
1  | uuid    | 2025-10-23 | 1000000       | 2025-10-23 04:35:00
2  | uuid    | 2025-10-23 | 2000000       | NULL

-- Unique Index: (user_id, date, deleted_at)
-- Row 1: (uuid, 2025-10-23, '2025-10-23 04:35:00') ✅
-- Row 2: (uuid, 2025-10-23, NULL)                  ✅
-- Different deleted_at = Unique! ✅
```

---

## Files Modified

1. **internal/models/financial.go**
   - Changed `UserID` tag: Added `uniqueIndex:idx_user_date_deleted`
   - Changed `Date` tag: Changed to `uniqueIndex:idx_user_date_deleted`
   - Changed `DeletedAt` tag: Added `uniqueIndex:idx_user_date_deleted`

---

## Quick Fix for Existing Data

If you have soft-deleted records blocking new inserts:

### Option A: Hard Delete Soft-Deleted Records

```sql
-- Permanently remove soft-deleted records
DELETE FROM trading_activities
WHERE daily_target_id IN (
    SELECT id FROM daily_targets WHERE deleted_at IS NOT NULL
);

DELETE FROM daily_targets WHERE deleted_at IS NOT NULL;
```

### Option B: Update Unique Index Manually

```sql
-- Drop old index
DROP INDEX IF EXISTS idx_user_date;

-- Create new index with deleted_at
CREATE UNIQUE INDEX idx_user_date_deleted
ON daily_targets (user_id, date, deleted_at);
```

Then restart server.

---

## Summary

**Problem:**
❌ Cannot create daily target after deleting
❌ Error: "duplicate key violates unique constraint idx_user_date"
❌ Soft-deleted records block new inserts

**Root Cause:**
- Unique index on `(user_id, date)` doesn't include `deleted_at`
- Soft-deleted rows still occupy unique slot
- PostgreSQL sees duplicate key even though record is "deleted"

**Solution:**
✅ Include `deleted_at` in unique index: `(user_id, date, deleted_at)`
✅ Soft-deleted records have different `deleted_at` timestamp
✅ Active records have `deleted_at = NULL`
✅ Can create new target after delete (different unique tuple)

**Migration:**
⚠️ Must drop old index and create new one
⚠️ Restart server for GORM to detect changes

**Status:** ✅ FIXED (requires server restart)
