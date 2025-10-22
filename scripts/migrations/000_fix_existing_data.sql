-- Quick Fix: Manually delete orphaned data and fix constraint
-- Run this BEFORE the main migration if you have existing data issues

-- Step 1: Find daily targets that have trading activities
SELECT
    dt.id AS target_id,
    dt.date,
    dt.user_id,
    COUNT(ta.id) AS trading_activities_count
FROM daily_targets dt
LEFT JOIN trading_activities ta ON ta.daily_target_id = dt.id
WHERE dt.deleted_at IS NULL
GROUP BY dt.id, dt.date, dt.user_id
HAVING COUNT(ta.id) > 0
ORDER BY dt.date DESC;

-- Step 2: Choose your approach:

-- OPTION A: Keep everything, just fix the constraint (RECOMMENDED)
-- No data deletion needed, just run the main migration script
\i 001_add_cascade_delete_trading_activities.sql

-- OPTION B: Delete specific problematic targets WITH their activities
-- Replace target_id with the actual IDs you want to delete
BEGIN;

-- Delete trading activities first (to satisfy current constraint)
DELETE FROM trading_activities
WHERE daily_target_id IN (2, 3);  -- Replace with your target IDs

-- Then delete the daily targets
DELETE FROM daily_targets
WHERE id IN (2, 3);  -- Replace with your target IDs

COMMIT;

-- OPTION C: Clean ALL test data and start fresh
BEGIN;

-- Show what will be deleted
SELECT 'Daily Targets to delete:' AS info, COUNT(*) FROM daily_targets;
SELECT 'Trading Activities to delete:' AS info, COUNT(*) FROM trading_activities;

-- Uncomment to execute:
-- DELETE FROM trading_activities;  -- Delete all trading activities
-- DELETE FROM daily_targets;       -- Delete all daily targets

COMMIT;

-- Step 3: Verify cleanup
SELECT
    'daily_targets' AS table_name,
    COUNT(*) AS row_count
FROM daily_targets
WHERE deleted_at IS NULL
UNION ALL
SELECT
    'trading_activities' AS table_name,
    COUNT(*) AS row_count
FROM trading_activities
WHERE deleted_at IS NULL;

-- Step 4: Now run the main migration
-- \i 001_add_cascade_delete_trading_activities.sql
