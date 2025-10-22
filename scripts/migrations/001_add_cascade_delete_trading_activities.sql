-- Migration: Add CASCADE DELETE to trading_activities foreign key
-- Date: 2025-10-23
-- Purpose: Fix foreign key constraint violation when deleting daily targets

-- Step 1: Check existing constraint
SELECT
    tc.constraint_name,
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name,
    rc.delete_rule
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
LEFT JOIN information_schema.referential_constraints AS rc
    ON rc.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_name = 'trading_activities'
    AND kcu.column_name = 'daily_target_id';

-- Expected output:
-- constraint_name: fk_daily_targets_trading_activities
-- delete_rule: NO ACTION (or RESTRICT)

-- Step 2: Drop existing foreign key constraint
ALTER TABLE trading_activities
DROP CONSTRAINT IF EXISTS fk_daily_targets_trading_activities;

-- Step 3: Add new foreign key constraint WITH CASCADE DELETE
ALTER TABLE trading_activities
ADD CONSTRAINT fk_daily_targets_trading_activities
FOREIGN KEY (daily_target_id)
REFERENCES daily_targets(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- Step 4: Verify the new constraint
SELECT
    tc.constraint_name,
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name,
    rc.delete_rule,
    rc.update_rule
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
LEFT JOIN information_schema.referential_constraints AS rc
    ON rc.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_name = 'trading_activities'
    AND kcu.column_name = 'daily_target_id';

-- Expected output:
-- constraint_name: fk_daily_targets_trading_activities
-- delete_rule: CASCADE
-- update_rule: CASCADE

-- Step 5: Test CASCADE DELETE (optional)
-- Create test data
DO $$
DECLARE
    test_user_id UUID := '00000000-0000-0000-0000-000000000001';
    test_target_id INT;
BEGIN
    -- Insert test daily target
    INSERT INTO daily_targets (user_id, date, income_target, expense_limit, remaining_income, remaining_expense)
    VALUES (test_user_id, CURRENT_DATE, 1000000, 100000, 1000000, 100000)
    RETURNING id INTO test_target_id;

    -- Insert test trading activities
    INSERT INTO trading_activities (daily_target_id, user_id, trade_type, amount, symbol, trade_time)
    VALUES
        (test_target_id, test_user_id, 'win', 100000, 'BTCUSD', NOW()),
        (test_target_id, test_user_id, 'loss', 50000, 'EURUSD', NOW());

    RAISE NOTICE 'Test data created: daily_target_id=%', test_target_id;

    -- Verify trading activities exist
    IF (SELECT COUNT(*) FROM trading_activities WHERE daily_target_id = test_target_id) != 2 THEN
        RAISE EXCEPTION 'Trading activities not created correctly';
    END IF;

    -- Delete the daily target (CASCADE should delete trading activities)
    DELETE FROM daily_targets WHERE id = test_target_id;

    -- Verify trading activities were deleted
    IF (SELECT COUNT(*) FROM trading_activities WHERE daily_target_id = test_target_id) != 0 THEN
        RAISE EXCEPTION 'CASCADE DELETE failed - trading activities still exist';
    END IF;

    RAISE NOTICE 'CASCADE DELETE test passed successfully';
END $$;

-- Cleanup test user (optional)
-- DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';

-- Migration complete!
RAISE NOTICE 'Migration completed successfully: CASCADE DELETE enabled for trading_activities';
