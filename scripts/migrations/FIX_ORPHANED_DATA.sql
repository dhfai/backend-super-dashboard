-- Emergency Fix: Clean orphaned trading activities
-- Run this BEFORE starting the server

-- Step 1: Find orphaned trading activities (activities without valid daily_target)
SELECT
    ta.id,
    ta.daily_target_id,
    ta.user_id,
    ta.trade_type,
    ta.amount,
    ta.created_at
FROM trading_activities ta
LEFT JOIN daily_targets dt ON dt.id = ta.daily_target_id
WHERE dt.id IS NULL
ORDER BY ta.created_at DESC;

-- Step 2: Count orphaned records
SELECT COUNT(*) AS orphaned_activities
FROM trading_activities ta
LEFT JOIN daily_targets dt ON dt.id = ta.daily_target_id
WHERE dt.id IS NULL;

-- Step 3: DELETE orphaned trading activities
BEGIN;

DELETE FROM trading_activities
WHERE id IN (
    SELECT ta.id
    FROM trading_activities ta
    LEFT JOIN daily_targets dt ON dt.id = ta.daily_target_id
    WHERE dt.id IS NULL
);

-- Show how many deleted
SELECT 'Deleted orphaned trading activities' AS status;

COMMIT;

-- Step 4: Verify no orphans remain
SELECT COUNT(*) AS remaining_orphans
FROM trading_activities ta
LEFT JOIN daily_targets dt ON dt.id = ta.daily_target_id
WHERE dt.id IS NULL;
-- Expected: 0

-- Step 5: Now you can safely start the server
-- The foreign key constraint will be created successfully
