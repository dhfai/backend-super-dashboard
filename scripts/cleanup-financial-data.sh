#!/bin/bash

# Clean up test data dari database

echo "🧹 Cleaning up test data..."

# Pastikan Anda sudah set environment variables untuk database
# atau edit langsung connection string di bawah

PGPASSWORD="${DB_PASSWORD:-password}" psql \
  -h "${DB_HOST:-localhost}" \
  -p "${DB_PORT:-5432}" \
  -U "${DB_USER:-postgres}" \
  -d "${DB_NAME:-filestore}" \
  -c "
    -- Delete trading activities first (foreign key constraint)
    DELETE FROM trading_activities;

    -- Delete daily targets
    DELETE FROM daily_targets;

    -- Delete transactions
    DELETE FROM transactions;

    -- Delete financial goals
    DELETE FROM financial_goals;

    -- Delete backtest strategies
    DELETE FROM backtest_strategies;

    -- Delete budgets
    DELETE FROM budgets;

    -- Show counts
    SELECT
      (SELECT COUNT(*) FROM trading_activities) as trading_activities,
      (SELECT COUNT(*) FROM daily_targets) as daily_targets,
      (SELECT COUNT(*) FROM transactions) as transactions,
      (SELECT COUNT(*) FROM financial_goals) as financial_goals,
      (SELECT COUNT(*) FROM backtest_strategies) as backtest_strategies,
      (SELECT COUNT(*) FROM budgets) as budgets;
  "

echo "✅ Cleanup completed!"
