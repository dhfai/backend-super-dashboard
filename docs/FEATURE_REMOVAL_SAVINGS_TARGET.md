# Removal of Savings Target Field

## Date: October 24, 2025

## Summary
Removed the `savings_target` field from the Daily Target feature as it's no longer needed in the form.

## Changes Made

### 1. Model Changes (`internal/models/financial.go`)
- Removed `SavingsTarget` field from `DailyTarget` struct

### 2. DTO Changes (`internal/models/dto.go`)
- Removed `SavingsTarget` field from `CreateDailyTargetRequest`
- Removed `SavingsTarget` field from `UpdateDailyTargetRequest`
- Removed `SavingsTarget` field from `DailyTargetResponse`
- Removed `SavingsProgress` field from `DailyTargetResponse`
- Updated `ToDailyTargetResponse()` method to remove savings-related calculations

### 3. Service Changes (`internal/services/daily_target_service.go`)
- Removed `SavingsTarget` assignment in `CreateDailyTarget()`
- Removed `SavingsTarget` update logic in `UpdateDailyTarget()`
- Removed `total_savings_target` and `days_met_savings` from summary calculations in `GetDailyTargetsSummary()`

### 4. Database Migration
- Created migration file: `scripts/migrations/002_remove_savings_target.sql`
- Migration removes the `savings_target` column from `daily_targets` table

## API Changes

### Create Daily Target Request
**Before:**
```json
{
  "date": "2025-10-24",
  "income_target": 5000000,
  "expense_limit": 500000,
  "savings_target": 1000000,
  "notes": "Trading plan..."
}
```

**After:**
```json
{
  "date": "2025-10-24",
  "income_target": 5000000,
  "expense_limit": 500000,
  "notes": "Trading plan..."
}
```

### Update Daily Target Request
**Before:**
```json
{
  "income_target": 6000000,
  "expense_limit": 600000,
  "savings_target": 1200000,
  "notes": "Updated plan..."
}
```

**After:**
```json
{
  "income_target": 6000000,
  "expense_limit": 600000,
  "notes": "Updated plan..."
}
```

### Daily Target Response
**Before:**
```json
{
  "id": 1,
  "income_target": 5000000,
  "expense_limit": 500000,
  "savings_target": 1000000,
  "savings_progress": 75.5,
  ...
}
```

**After:**
```json
{
  "id": 1,
  "income_target": 5000000,
  "expense_limit": 500000,
  ...
}
```

## Migration Instructions

To apply this change to an existing database, run the migration script:

```bash
psql -U your_username -d your_database -f scripts/migrations/002_remove_savings_target.sql
```

Or if using the application's auto-migration feature, the column will be automatically removed on the next application restart.

## Impact
- **Breaking Change**: Yes - API consumers need to remove `savings_target` from their requests
- **Database Migration**: Required
- **Frontend Changes**: Required - Remove the "Savings Target" input field from the form

## Testing
After deployment, verify:
1. ✅ Can create daily targets without `savings_target` field
2. ✅ Can update daily targets without `savings_target` field
3. ✅ Response no longer includes `savings_target` or `savings_progress`
4. ✅ Summary endpoint no longer includes `total_savings_target` or `days_met_savings`
5. ✅ Database column `savings_target` has been removed

## Rollback
If needed, to rollback this change:
1. Restore the removed code from git history
2. Run the following SQL:
   ```sql
   ALTER TABLE daily_targets ADD COLUMN savings_target DECIMAL(15,2) NOT NULL DEFAULT 0;
   ```
