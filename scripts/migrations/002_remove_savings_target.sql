-- Migration: Remove savings_target column from daily_targets table
-- Date: 2025-10-24
-- Description: Removes the savings_target field as it's no longer needed in the daily target form

-- Drop the savings_target column
ALTER TABLE daily_targets DROP COLUMN IF EXISTS savings_target;
