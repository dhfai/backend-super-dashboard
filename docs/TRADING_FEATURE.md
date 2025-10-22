# Trading Feature - Daily Target untuk Trading

## Overview

Daily Target kini sudah ditingkatkan untuk **Trading Performance Tracking** dengan fitur real-time tracking profit/loss dan auto-completion ketika target tercapai atau loss limit exceeded.

## Konsep

### Daily Target untuk Trading
- **Income Target**: Target profit yang ingin dicapai (misal: Rp 1,000,000)
- **Expense Limit**: Maksimal loss yang diperbolehkan (stop loss, misal: Rp 100,000)
- **Savings Target**: Optional, net profit target (income - expense)

### Trading Activity
Setiap trade (win/loss) dicatat sebagai **Trading Activity** yang otomatis:
- ✅ **Mengurangi Income Target** jika win (profit)
- ✅ **Mengurangi Expense Limit** jika loss
- ✅ **Menghitung Win Rate** otomatis
- ✅ **Auto-complete** ketika target tercapai atau loss limit exceeded

## Database Schema

### DailyTarget (Updated)
```sql
CREATE TABLE daily_targets (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    date DATE NOT NULL,
    income_target DECIMAL(15,2) NOT NULL,
    expense_limit DECIMAL(15,2) NOT NULL,
    savings_target DECIMAL(15,2) NOT NULL,
    actual_income DECIMAL(15,2) DEFAULT 0,
    actual_expense DECIMAL(15,2) DEFAULT 0,
    actual_savings DECIMAL(15,2) DEFAULT 0,
    remaining_income DECIMAL(15,2) DEFAULT 0,
    remaining_expense DECIMAL(15,2) DEFAULT 0,
    total_trades INT DEFAULT 0,
    winning_trades INT DEFAULT 0,
    losing_trades INT DEFAULT 0,
    win_rate DECIMAL(5,2) DEFAULT 0,
    is_completed BOOLEAN DEFAULT false,
    completed_at TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(user_id, date)
);
```

### TradingActivity (New)
```sql
CREATE TABLE trading_activities (
    id SERIAL PRIMARY KEY,
    daily_target_id INT NOT NULL REFERENCES daily_targets(id),
    user_id UUID NOT NULL,
    trade_type VARCHAR(20) NOT NULL,  -- 'win' or 'loss'
    amount DECIMAL(15,2) NOT NULL,
    pips INT DEFAULT 0,
    lot_size DECIMAL(10,2) DEFAULT 0,
    symbol VARCHAR(50),  -- e.g., EURUSD
    description TEXT,
    trade_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## API Endpoints

### 1. Create Daily Target (Trading Plan)

**Endpoint:** `POST /api/v1/finance/daily-targets`

**Request:**
```json
{
  "date": "2025-10-20T00:00:00Z",
  "income_target": 1000000.00,
  "expense_limit": 100000.00,
  "savings_target": 50000.00,
  "notes": "Trading Plan Hari Senin"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Daily target created successfully",
  "data": {
    "id": 1,
    "user_id": "uuid",
    "date": "2025-10-20T00:00:00Z",
    "income_target": 1000000.00,
    "expense_limit": 100000.00,
    "savings_target": 50000.00,
    "actual_income": 0.00,
    "actual_expense": 0.00,
    "actual_savings": 0.00,
    "remaining_income": 1000000.00,
    "remaining_expense": 100000.00,
    "total_trades": 0,
    "winning_trades": 0,
    "losing_trades": 0,
    "win_rate": 0.00,
    "is_completed": false,
    "notes": "Trading Plan Hari Senin"
  }
}
```

### 2. Add Trading Activity (Record Win/Loss)

**Endpoint:** `POST /api/v1/finance/daily-targets/:targetId/trades`

**Request - Profit (Win):**
```json
{
  "trade_type": "win",
  "amount": 160000.00,
  "pips": 100,
  "lot_size": 0.01,
  "symbol": "EURUSD",
  "description": "TP hit, bullish breakout",
  "trade_time": "2025-10-20T09:30:00Z"
}
```

**Request - Loss:**
```json
{
  "trade_type": "loss",
  "amount": 30000.00,
  "pips": 30,
  "lot_size": 0.01,
  "symbol": "GBPUSD",
  "description": "SL hit, reversal",
  "trade_time": "2025-10-20T11:15:00Z"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Trade added successfully",
  "data": {
    "activity": {
      "id": 1,
      "daily_target_id": 1,
      "user_id": "uuid",
      "trade_type": "win",
      "amount": 160000.00,
      "pips": 100,
      "lot_size": 0.01,
      "symbol": "EURUSD",
      "description": "TP hit, bullish breakout",
      "trade_time": "2025-10-20T09:30:00Z"
    },
    "target": {
      "id": 1,
      "income_target": 1000000.00,
      "expense_limit": 100000.00,
      "actual_income": 160000.00,
      "actual_expense": 0.00,
      "actual_savings": 160000.00,
      "remaining_income": 840000.00,
      "remaining_expense": 100000.00,
      "total_trades": 1,
      "winning_trades": 1,
      "losing_trades": 0,
      "win_rate": 100.00,
      "is_completed": false,
      "income_progress": 16.00,
      "expense_progress": 0.00
    }
  }
}
```

### 3. Get Trading Activities by Daily Target

**Endpoint:** `GET /api/v1/finance/daily-targets/:targetId/trades`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Trading activities retrieved successfully",
  "data": [
    {
      "id": 1,
      "daily_target_id": 1,
      "trade_type": "win",
      "amount": 160000.00,
      "pips": 100,
      "lot_size": 0.01,
      "symbol": "EURUSD",
      "trade_time": "2025-10-20T09:30:00Z"
    },
    {
      "id": 2,
      "daily_target_id": 1,
      "trade_type": "loss",
      "amount": 30000.00,
      "pips": 30,
      "lot_size": 0.01,
      "symbol": "GBPUSD",
      "trade_time": "2025-10-20T11:15:00Z"
    }
  ]
}
```

### 4. Get Trading Stats

**Endpoint:** `GET /api/v1/finance/trading-activities/stats?date_from=2025-10-01&date_to=2025-10-31`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Trading stats retrieved successfully",
  "data": {
    "total_trades": 25,
    "winning_trades": 18,
    "losing_trades": 7,
    "win_rate": 72.00,
    "total_profit": 3500000.00,
    "total_loss": 450000.00,
    "net_profit": 3050000.00,
    "date_from": "2025-10-01T00:00:00Z",
    "date_to": "2025-10-31T23:59:59Z"
  }
}
```

### 5. Delete Trading Activity

**Endpoint:** `DELETE /api/v1/finance/trading-activities/:id`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Trading activity deleted successfully"
}
```

**Note:** Deleting a trading activity will automatically recalculate the daily target stats (actual_income, actual_expense, win_rate, etc.)

## Business Logic

### Auto-Calculation saat Add Trade

#### When Trade Type = "win" (Profit)
```go
target.ActualIncome += amount
target.WinningTrades++
target.TotalTrades++
target.ActualSavings = target.ActualIncome - target.ActualExpense
target.RemainingIncome = target.IncomeTarget - target.ActualIncome
target.WinRate = (WinningTrades / TotalTrades) * 100

// Auto-complete if income target reached
if target.ActualIncome >= target.IncomeTarget {
    target.IsCompleted = true
    target.CompletedAt = now
}
```

#### When Trade Type = "loss"
```go
target.ActualExpense += amount
target.LosingTrades++
target.TotalTrades++
target.ActualSavings = target.ActualIncome - target.ActualExpense
target.RemainingExpense = target.ExpenseLimit - target.ActualExpense
target.WinRate = (WinningTrades / TotalTrades) * 100

// Auto-complete if expense limit exceeded (stop loss)
if target.ActualExpense >= target.ExpenseLimit {
    target.IsCompleted = true
    target.CompletedAt = now
}
```

## Use Case Examples

### Scenario 1: Trading Harian Normal

**Setup:**
```json
{
  "date": "2025-10-20",
  "income_target": 1000000,
  "expense_limit": 100000,
  "notes": "Scalping EURUSD & GBPUSD"
}
```

**Trading Activities:**
1. Win: +160,000 (100 pips, 0.01 lot) → Remaining: 840,000
2. Win: +150,000 (95 pips, 0.01 lot) → Remaining: 690,000
3. Loss: -30,000 (30 pips SL) → Remaining Expense: 70,000
4. Win: +180,000 (110 pips) → Remaining: 510,000
5. ... (continue until target reached or day ends)

### Scenario 2: Target Tercapai

```json
// After 7 trades:
{
  "actual_income": 1050000,
  "actual_expense": 45000,
  "total_trades": 7,
  "winning_trades": 6,
  "losing_trades": 1,
  "win_rate": 85.71,
  "is_completed": true,  // ✅ Auto-completed
  "completed_at": "2025-10-20T14:30:00Z"
}
```

### Scenario 3: Stop Loss Hit

```json
// After 10 trades with bad day:
{
  "actual_income": 300000,
  "actual_expense": 105000,  // Exceeded limit!
  "total_trades": 10,
  "winning_trades": 4,
  "losing_trades": 6,
  "win_rate": 40.00,
  "is_completed": true,  // ✅ Auto-stopped
  "completed_at": "2025-10-20T16:00:00Z"
}
```

## Frontend Integration

### Create Daily Target
```javascript
POST /api/v1/finance/daily-targets
{
  "date": "2025-10-20T00:00:00Z",
  "income_target": 1000000,
  "expense_limit": 100000,
  "savings_target": 50000,
  "notes": "Trading Plan"
}
```

### Add Trade (Win)
```javascript
POST /api/v1/finance/daily-targets/1/trades
{
  "trade_type": "win",
  "amount": 160000,
  "pips": 100,
  "lot_size": 0.01,
  "symbol": "EURUSD"
}
```

### Add Trade (Loss)
```javascript
POST /api/v1/finance/daily-targets/1/trades
{
  "trade_type": "loss",
  "amount": 30000,
  "pips": 30,
  "lot_size": 0.01,
  "symbol": "GBPUSD"
}
```

### Get Today's Target with Trades
```javascript
GET /api/v1/finance/daily-targets/today
// Response includes trading_activities array
```

## Benefits

1. ✅ **Real-time Tracking**: Setiap trade langsung update target
2. ✅ **Auto-calculation**: Win rate, remaining, progress otomatis dihitung
3. ✅ **Auto-completion**: Target tercapai atau stop loss hit → otomatis complete
4. ✅ **Statistics**: Win rate, total profit/loss, net profit
5. ✅ **Trading Journal**: Semua trade history tersimpan (pips, lot size, symbol)
6. ✅ **Risk Management**: Expense limit untuk stop loss protection

## Testing

See `tests/trading_api.http` for complete test cases.

## Migration Notes

Jika Anda sudah punya daily targets lama, fields baru akan otomatis diisi dengan default values:
- `remaining_income` = `income_target`
- `remaining_expense` = `expense_limit`
- `total_trades`, `winning_trades`, `losing_trades` = 0
- `win_rate` = 0.00
- `is_completed` = false
