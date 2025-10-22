# Financial Management API Documentation

## Overview

The Financial Management API provides comprehensive endpoints for managing personal finances including transactions, budgets, financial goals, daily targets, and backtest strategies.

**Base URL:** `http://localhost:8080/api/v1/finance`

**Authentication:** All endpoints require JWT authentication via `Authorization: Bearer <token>` header.

---

## Transactions API

Manage income and expense transactions.

### 1. Create Transaction

Create a new financial transaction.

**Endpoint:** `POST /finance/transactions`

**Request Body:**
```json
{
  "type": "income",
  "amount": 5000.00,
  "category": "Salary",
  "description": "Monthly salary",
  "date": "2025-10-15T00:00:00Z",
  "tags": "monthly,income,salary"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Transaction created successfully",
  "data": {
    "id": 1,
    "user_id": "uuid",
    "type": "income",
    "amount": 5000.00,
    "category": "Salary",
    "description": "Monthly salary",
    "date": "2025-10-15T00:00:00Z",
    "tags": "monthly,income,salary",
    "created_at": "2025-10-19T10:00:00Z",
    "updated_at": "2025-10-19T10:00:00Z"
  }
}
```

### 2. Get All Transactions

Retrieve all transactions with filters and pagination.

**Endpoint:** `GET /finance/transactions`

**Query Parameters:**
- `type` (string, optional): Filter by type (`income` or `expense`)
- `category` (string, optional): Filter by category
- `tags` (string, optional): Filter by tags
- `date_from` (date, optional): Start date (YYYY-MM-DD)
- `date_to` (date, optional): End date (YYYY-MM-DD)
- `min_amount` (number, optional): Minimum amount
- `max_amount` (number, optional): Maximum amount
- `page` (int, optional): Page number (default: 1)
- `page_size` (int, optional): Items per page (default: 10)
- `sort_by` (string, optional): Sort field (`date`, `amount`, `created_at`)
- `sort_order` (string, optional): Sort order (`asc`, `desc`)

**Example:** `GET /finance/transactions?type=expense&date_from=2025-10-01&date_to=2025-10-31&page=1&page_size=20`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Transactions retrieved successfully",
  "data": {
    "transactions": [...],
    "total": 50,
    "page": 1,
    "page_size": 20,
    "total_pages": 3,
    "summary": {
      "total_income": 5000.00,
      "total_expense": 3000.00,
      "net_balance": 2000.00
    }
  }
}
```

### 3. Get Transaction by ID

**Endpoint:** `GET /finance/transactions/:id`

**Response:** `200 OK`

### 4. Update Transaction

**Endpoint:** `PUT /finance/transactions/:id`

**Request Body:**
```json
{
  "amount": 5500.00,
  "description": "Updated description"
}
```

### 5. Delete Transaction

**Endpoint:** `DELETE /finance/transactions/:id`

**Response:** `200 OK`

### 6. Get Transactions by Category

**Endpoint:** `GET /finance/transactions/category/:category`

**Query Parameters:**
- `page` (int, optional)
- `page_size` (int, optional)

### 7. Get Transaction Summary

Get income/expense summary for a period.

**Endpoint:** `GET /finance/transactions/summary`

**Query Parameters:**
- `date_from` (date, optional)
- `date_to` (date, optional)

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Summary retrieved successfully",
  "data": {
    "total_income": 15000.00,
    "total_expense": 8500.00,
    "net_balance": 6500.00
  }
}
```

### 8. Get Category Breakdown

**Endpoint:** `GET /finance/transactions/breakdown`

**Query Parameters:**
- `type` (string, optional): `income` or `expense`
- `date_from` (date, optional)
- `date_to` (date, optional)

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Breakdown retrieved successfully",
  "data": {
    "Food": 1500.00,
    "Transport": 800.00,
    "Entertainment": 500.00,
    "Bills": 2000.00
  }
}
```

### 9. Get Monthly Trend

**Endpoint:** `GET /finance/transactions/trend`

**Query Parameters:**
- `months` (int, optional): Number of months (default: 6)

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Trend retrieved successfully",
  "data": [
    {
      "month": "2025-05",
      "income": 5000.00,
      "expense": 3200.00,
      "net_balance": 1800.00
    },
    ...
  ]
}
```

---

## Daily Targets API

Track and manage daily financial targets.

### 1. Create Daily Target

**Endpoint:** `POST /finance/daily-targets`

**Request Body:**
```json
{
  "date": "2025-10-20T00:00:00Z",
  "income_target": 500.00,
  "expense_limit": 200.00,
  "savings_target": 300.00,
  "notes": "Weekend target"
}
```

**Response:** `201 Created`

### 2. Get All Daily Targets

**Endpoint:** `GET /finance/daily-targets`

**Query Parameters:**
- `date_from` (date, optional)
- `date_to` (date, optional)
- `page` (int, optional)
- `page_size` (int, optional)

### 3. Get Today's Target

**Endpoint:** `GET /finance/daily-targets/today`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Today's target retrieved successfully",
  "data": {
    "id": 1,
    "user_id": "uuid",
    "date": "2025-10-19T00:00:00Z",
    "income_target": 500.00,
    "expense_limit": 200.00,
    "savings_target": 300.00,
    "actual_income": 450.00,
    "actual_expense": 180.00,
    "actual_savings": 270.00,
    "income_progress": 90.00,
    "expense_progress": 90.00,
    "savings_progress": 90.00,
    "notes": "Weekend target"
  }
}
```

### 4. Get Current Month Summary

**Endpoint:** `GET /finance/daily-targets/month-summary`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Month summary retrieved successfully",
  "data": {
    "total_days": 15,
    "total_income_target": 7500.00,
    "total_expense_limit": 3000.00,
    "total_savings_target": 4500.00,
    "total_actual_income": 7200.00,
    "total_actual_expense": 2800.00,
    "total_actual_savings": 4400.00,
    "days_met_income": 12,
    "days_met_expense": 14,
    "days_met_savings": 11
  }
}
```

### 5. Get Week Summary

**Endpoint:** `GET /finance/daily-targets/week-summary`

### 6. Get Daily Target by ID

**Endpoint:** `GET /finance/daily-targets/:id`

### 7. Update Daily Target

**Endpoint:** `PUT /finance/daily-targets/:id`

### 8. Delete Daily Target

**Endpoint:** `DELETE /finance/daily-targets/:id`

### 9. Refresh Actual Values

Recalculate actual values from transactions.

**Endpoint:** `POST /finance/daily-targets/:id/refresh`

---

## Financial Goals API

Manage long-term financial goals.

### 1. Create Financial Goal

**Endpoint:** `POST /finance/goals`

**Request Body:**
```json
{
  "title": "Emergency Fund",
  "description": "Build 6 months emergency fund",
  "target_amount": 30000.00,
  "current_amount": 5000.00,
  "start_date": "2025-01-01T00:00:00Z",
  "target_date": "2025-12-31T00:00:00Z",
  "category": "Savings",
  "priority": 2
}
```

**Priority Values:**
- `0`: Low
- `1`: Medium
- `2`: High

**Response:** `201 Created`

### 2. Get All Financial Goals

**Endpoint:** `GET /finance/goals`

**Query Parameters:**
- `category` (string, optional)
- `priority` (int, optional): 0, 1, or 2
- `is_completed` (bool, optional)
- `page` (int, optional)
- `page_size` (int, optional)

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Financial goals retrieved successfully",
  "data": {
    "goals": [
      {
        "id": 1,
        "user_id": "uuid",
        "title": "Emergency Fund",
        "description": "Build 6 months emergency fund",
        "target_amount": 30000.00,
        "current_amount": 15000.00,
        "progress": 50.00,
        "start_date": "2025-01-01T00:00:00Z",
        "target_date": "2025-12-31T00:00:00Z",
        "category": "Savings",
        "priority": 2,
        "priority_label": "high",
        "is_completed": false,
        "days_remaining": 73,
        "required_per_day": 205.48
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 10,
    "total_pages": 1
  }
}
```

### 3. Get Financial Goal by ID

**Endpoint:** `GET /finance/goals/:id`

### 4. Update Financial Goal

**Endpoint:** `PUT /finance/goals/:id`

### 5. Delete Financial Goal

**Endpoint:** `DELETE /finance/goals/:id`

### 6. Update Goal Progress

Set current amount to a specific value.

**Endpoint:** `POST /finance/goals/:id/update-progress`

**Request Body:**
```json
{
  "amount": 18000.00
}
```

### 7. Add to Goal Progress

Add amount to current progress.

**Endpoint:** `POST /finance/goals/:id/add-progress`

**Request Body:**
```json
{
  "amount": 1000.00
}
```

### 8. Get Goals Summary

**Endpoint:** `GET /finance/goals/summary`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Goals summary retrieved successfully",
  "data": {
    "total_goals": 5,
    "active_goals": 3,
    "completed_goals": 2,
    "overdue_goals": 1,
    "total_target": 100000.00,
    "total_saved": 45000.00,
    "total_remaining": 55000.00,
    "overall_progress": 45.00,
    "high_priority": 2,
    "medium_priority": 2,
    "low_priority": 1
  }
}
```

---

## Backtest Strategy API

Create and analyze investment/savings strategies.

### 1. Create Backtest Strategy

**Endpoint:** `POST /finance/backtest-strategies`

**Request Body:**
```json
{
  "name": "Monthly Investment Plan",
  "description": "Invest $500 monthly with 8% annual return",
  "strategy_type": "investment",
  "initial_amount": 10000.00,
  "monthly_contribution": 500.00,
  "expected_return": 8.0,
  "duration": 60,
  "start_date": "2025-01-01T00:00:00Z"
}
```

**Strategy Types:**
- `savings`: Simple savings strategy
- `investment`: Investment with returns
- `expense_reduction`: Expense reduction strategy

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Backtest strategy created successfully",
  "data": {
    "id": 1,
    "user_id": "uuid",
    "name": "Monthly Investment Plan",
    "description": "Invest $500 monthly with 8% annual return",
    "strategy_type": "investment",
    "initial_amount": 10000.00,
    "monthly_contribution": 500.00,
    "expected_return": 8.0,
    "duration": 60,
    "start_date": "2025-01-01T00:00:00Z",
    "end_date": "2030-01-01T00:00:00Z",
    "projected_amount": 73476.85,
    "actual_amount": 0.00,
    "status": "draft",
    "performance_gap": 0.00
  }
}
```

### 2. Get All Backtest Strategies

**Endpoint:** `GET /finance/backtest-strategies`

**Query Parameters:**
- `strategy_type` (string, optional)
- `status` (string, optional): `draft`, `active`, `completed`, `cancelled`
- `page` (int, optional)
- `page_size` (int, optional)

### 3. Get Backtest Strategy by ID

**Endpoint:** `GET /finance/backtest-strategies/:id`

### 4. Get Backtest Result

Get detailed calculation results.

**Endpoint:** `GET /finance/backtest-strategies/:id/result`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Backtest result retrieved successfully",
  "data": {
    "initial_amount": 10000.00,
    "monthly_contribution": 500.00,
    "expected_return": 8.0,
    "duration": 60,
    "total_contributed": 40000.00,
    "total_returns": 33476.85,
    "final_amount": 73476.85,
    "monthly_results": [
      {
        "month": 1,
        "start_balance": 10000.00,
        "contribution": 500.00,
        "returns": 70.00,
        "end_balance": 10570.00,
        "total_returns": 70.00
      },
      ...
    ]
  }
}
```

### 5. Update Backtest Strategy

**Endpoint:** `PUT /finance/backtest-strategies/:id`

### 6. Delete Backtest Strategy

**Endpoint:** `DELETE /finance/backtest-strategies/:id`

### 7. Run Backtest

Recalculate backtest with current parameters.

**Endpoint:** `POST /finance/backtest-strategies/:id/run`

### 8. Activate Strategy

Change status from draft to active.

**Endpoint:** `POST /finance/backtest-strategies/:id/activate`

### 9. Complete Strategy

Mark strategy as completed with actual amount.

**Endpoint:** `POST /finance/backtest-strategies/:id/complete`

**Request Body:**
```json
{
  "actual_amount": 75000.00
}
```

---

## Budgets API

Manage category-based budgets.

### 1. Create Budget

**Endpoint:** `POST /finance/budgets`

**Request Body:**
```json
{
  "category": "Food & Dining",
  "amount": 3000.00,
  "period": "monthly",
  "start_date": "2025-10-01T00:00:00Z",
  "end_date": "2025-10-31T23:59:59Z",
  "is_recurring": true
}
```

**Period Values:**
- `weekly`
- `monthly`
- `yearly`

**Response:** `201 Created`

### 2. Get All Budgets

**Endpoint:** `GET /finance/budgets`

**Query Parameters:**
- `category` (string, optional)
- `period` (string, optional)
- `is_active` (bool, optional)
- `page` (int, optional)
- `page_size` (int, optional)

### 3. Get Budget by ID

**Endpoint:** `GET /finance/budgets/:id`

### 4. Update Budget

**Endpoint:** `PUT /finance/budgets/:id`

### 5. Delete Budget

**Endpoint:** `DELETE /finance/budgets/:id`

### 6. Get Budget Summary

**Endpoint:** `GET /finance/budgets/summary`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Budget summary retrieved successfully",
  "data": {
    "total_budgets": 5,
    "total_allocated": 8000.00,
    "total_spent": 5500.00,
    "total_remaining": 2500.00,
    "over_budget_count": 1,
    "under_budget_count": 3,
    "on_track_count": 1,
    "categories": {
      "Food & Dining": {
        "allocated": 3000.00,
        "spent": 2500.00,
        "remaining": 500.00
      },
      ...
    }
  }
}
```

### 7. Get Monthly Budget Report

**Endpoint:** `GET /finance/budgets/monthly-report`

**Query Parameters:**
- `year` (int, optional): Default current year
- `month` (int, optional): Default current month

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Monthly budget report retrieved successfully",
  "data": {
    "year": 2025,
    "month": 10,
    "total_budgets": 5,
    "total_allocated": 8000.00,
    "total_spent": 5500.00,
    "total_remaining": 2500.00,
    "categories": [
      {
        "category": "Food & Dining",
        "allocated": 3000.00,
        "spent": 2800.00,
        "remaining": 200.00,
        "percentage": 93.33,
        "status": "critical"
      },
      ...
    ]
  }
}
```

### 8. Check Budget Alerts

Get budgets that exceed threshold.

**Endpoint:** `GET /finance/budgets/alerts`

**Query Parameters:**
- `threshold` (float, optional): Percentage threshold (default: 80)

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Budget alerts retrieved successfully",
  "data": [
    {
      "budget_id": 1,
      "category": "Food & Dining",
      "amount": 3000.00,
      "spent": 2800.00,
      "remaining": 200.00,
      "percentage": 93.33,
      "status": "critical"
    }
  ]
}
```

**Status Values:**
- `good`: < 75%
- `warning`: 75% - 89%
- `critical`: 90% - 99%
- `exceeded`: >= 100%

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "success": false,
  "message": "Invalid request data",
  "error": "validation error details"
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "message": "Unauthorized",
  "error": "invalid or missing token"
}
```

### 404 Not Found
```json
{
  "success": false,
  "message": "Resource not found",
  "error": "detailed error message"
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Internal server error",
  "error": "error details"
}
```

---

## Common Patterns

### Pagination

Most list endpoints support pagination:
- `page`: Page number (starts at 1)
- `page_size`: Items per page (default: 10, max: 100)

Response includes:
- `total`: Total number of items
- `page`: Current page
- `page_size`: Items per page
- `total_pages`: Total pages

### Filtering

Endpoints support various filters:
- Date ranges: `date_from`, `date_to`
- Amount ranges: `min_amount`, `max_amount`
- Categories, tags, types, status

### Sorting

Use `sort_by` and `sort_order` parameters:
- `sort_by`: Field name
- `sort_order`: `asc` or `desc`

---

## Best Practices

1. **Transactions**: Record transactions daily for accurate tracking
2. **Daily Targets**: Set realistic targets based on historical data
3. **Financial Goals**: Break large goals into smaller milestones
4. **Backtest Strategies**: Test multiple scenarios before committing
5. **Budgets**: Review and adjust budgets monthly
6. **Categories**: Use consistent category names across features
7. **Tags**: Leverage tags for flexible filtering and reporting

---

## Integration Examples

### Create Transaction and Update Daily Target

```bash
# 1. Create transaction
POST /api/v1/finance/transactions
{
  "type": "income",
  "amount": 500.00,
  "category": "Freelance",
  "date": "2025-10-19T00:00:00Z"
}

# 2. Refresh today's target
POST /api/v1/finance/daily-targets/:id/refresh
```

### Track Goal Progress with Transactions

```bash
# 1. Add savings transaction
POST /api/v1/finance/transactions
{
  "type": "income",
  "amount": 1000.00,
  "category": "Savings",
  "description": "Emergency fund contribution"
}

# 2. Update goal progress
POST /api/v1/finance/goals/:id/add-progress
{
  "amount": 1000.00
}
```

### Monitor Budget Status

```bash
# 1. Get budget alerts
GET /api/v1/finance/budgets/alerts?threshold=80

# 2. Get category breakdown
GET /api/v1/finance/transactions/breakdown?type=expense

# 3. Get monthly budget report
GET /api/v1/finance/budgets/monthly-report
```
