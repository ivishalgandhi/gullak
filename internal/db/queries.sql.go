// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const createTransaction = `-- name: CreateTransaction :many
INSERT INTO transactions (created_at, transaction_date, amount, currency, category, description, mode, confirm)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, created_at, transaction_date, currency, amount, category, mode, description, confirm
`

type CreateTransactionParams struct {
	CreatedAt       time.Time `json:"created_at"`
	TransactionDate time.Time `json:"transaction_date"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Category        string    `json:"category"`
	Description     string    `json:"description"`
	Mode            string    `json:"mode"`
	Confirm         bool      `json:"confirm"`
}

// Inserts a new transaction into the database.
func (q *Queries) CreateTransaction(ctx context.Context, arg CreateTransactionParams) ([]Transaction, error) {
	rows, err := q.query(ctx, q.createTransactionStmt, createTransaction,
		arg.CreatedAt,
		arg.TransactionDate,
		arg.Amount,
		arg.Currency,
		arg.Category,
		arg.Description,
		arg.Mode,
		arg.Confirm,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Transaction{}
	for rows.Next() {
		var i Transaction
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.TransactionDate,
			&i.Currency,
			&i.Amount,
			&i.Category,
			&i.Mode,
			&i.Description,
			&i.Confirm,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const dailySpending = `-- name: DailySpending :many

SELECT
    transaction_date,
    SUM(amount) AS total_spent
FROM transactions
WHERE transaction_date BETWEEN ? AND ?
GROUP BY transaction_date
ORDER BY transaction_date ASC
`

type DailySpendingParams struct {
	FromTransactionDate time.Time `json:"from_transaction_date"`
	ToTransactionDate   time.Time `json:"to_transaction_date"`
}

type DailySpendingRow struct {
	TransactionDate time.Time       `json:"transaction_date"`
	TotalSpent      sql.NullFloat64 `json:"total_spent"`
}

// Can be adjusted to show more or fewer categories
// Retrieves the sum total of all transactions for each day within a specified date range.
func (q *Queries) DailySpending(ctx context.Context, arg DailySpendingParams) ([]DailySpendingRow, error) {
	rows, err := q.query(ctx, q.dailySpendingStmt, dailySpending, arg.FromTransactionDate, arg.ToTransactionDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []DailySpendingRow{}
	for rows.Next() {
		var i DailySpendingRow
		if err := rows.Scan(&i.TransactionDate, &i.TotalSpent); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const deleteTransaction = `-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = ?
`

// Deletes a transaction by ID.
func (q *Queries) DeleteTransaction(ctx context.Context, id int64) error {
	_, err := q.exec(ctx, q.deleteTransactionStmt, deleteTransaction, id)
	return err
}

const getTransaction = `-- name: GetTransaction :one
SELECT id, created_at, transaction_date, currency, amount, category, mode, description, confirm FROM transactions WHERE id = ?
`

// Retrieves a single transaction by ID.
func (q *Queries) GetTransaction(ctx context.Context, id int64) (Transaction, error) {
	row := q.queryRow(ctx, q.getTransactionStmt, getTransaction, id)
	var i Transaction
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.TransactionDate,
		&i.Currency,
		&i.Amount,
		&i.Category,
		&i.Mode,
		&i.Description,
		&i.Confirm,
	)
	return i, err
}

const listTransactions = `-- name: ListTransactions :many
SELECT id, created_at, transaction_date, currency, amount, category, mode, description, confirm
FROM transactions
WHERE (?1 IS NULL OR confirm = ?1)
  AND (?2 IS NULL OR transaction_date >= ?2)
  AND (?3 IS NULL OR transaction_date <= ?3)
ORDER BY created_at DESC
`

type ListTransactionsParams struct {
	Confirm   interface{} `json:"confirm"`
	StartDate interface{} `json:"start_date"`
	EndDate   interface{} `json:"end_date"`
}

// Retrieves transactions optionally filtered by confirmation status and date range.
func (q *Queries) ListTransactions(ctx context.Context, arg ListTransactionsParams) ([]Transaction, error) {
	rows, err := q.query(ctx, q.listTransactionsStmt, listTransactions, arg.Confirm, arg.StartDate, arg.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Transaction{}
	for rows.Next() {
		var i Transaction
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.TransactionDate,
			&i.Currency,
			&i.Amount,
			&i.Category,
			&i.Mode,
			&i.Description,
			&i.Confirm,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const monthlySpendingSummary = `-- name: MonthlySpendingSummary :many
SELECT
    strftime('%Y', transaction_date) AS year,
    strftime('%m', transaction_date) AS month,
    category,
    SUM(amount) AS total_spent
FROM transactions
GROUP BY year, month, category
ORDER BY year DESC, month DESC, total_spent DESC
`

type MonthlySpendingSummaryRow struct {
	Year       interface{}     `json:"year"`
	Month      interface{}     `json:"month"`
	Category   string          `json:"category"`
	TotalSpent sql.NullFloat64 `json:"total_spent"`
}

// Returns the total spending grouped by month and category.
func (q *Queries) MonthlySpendingSummary(ctx context.Context) ([]MonthlySpendingSummaryRow, error) {
	rows, err := q.query(ctx, q.monthlySpendingSummaryStmt, monthlySpendingSummary)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MonthlySpendingSummaryRow{}
	for rows.Next() {
		var i MonthlySpendingSummaryRow
		if err := rows.Scan(
			&i.Year,
			&i.Month,
			&i.Category,
			&i.TotalSpent,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const topExpenseCategories = `-- name: TopExpenseCategories :many
SELECT
    category,
    SUM(amount) AS total_spent
FROM transactions
WHERE transaction_date BETWEEN ? AND ?  -- User specifies the start and end date
GROUP BY category
ORDER BY total_spent DESC
LIMIT 5
`

type TopExpenseCategoriesParams struct {
	FromTransactionDate time.Time `json:"from_transaction_date"`
	ToTransactionDate   time.Time `json:"to_transaction_date"`
}

type TopExpenseCategoriesRow struct {
	Category   string          `json:"category"`
	TotalSpent sql.NullFloat64 `json:"total_spent"`
}

// Retrieves the top expense categories over a specified period.
func (q *Queries) TopExpenseCategories(ctx context.Context, arg TopExpenseCategoriesParams) ([]TopExpenseCategoriesRow, error) {
	rows, err := q.query(ctx, q.topExpenseCategoriesStmt, topExpenseCategories, arg.FromTransactionDate, arg.ToTransactionDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TopExpenseCategoriesRow{}
	for rows.Next() {
		var i TopExpenseCategoriesRow
		if err := rows.Scan(&i.Category, &i.TotalSpent); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTransaction = `-- name: UpdateTransaction :exec
UPDATE transactions
SET amount = ?, currency = ?, category = ?, description = ?, mode = ?, confirm = ?, transaction_date = ?
WHERE id = ?
`

type UpdateTransactionParams struct {
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Category        string    `json:"category"`
	Description     string    `json:"description"`
	Mode            string    `json:"mode"`
	Confirm         bool      `json:"confirm"`
	TransactionDate time.Time `json:"transaction_date"`
	ID              int64     `json:"id"`
}

// Updates a transaction by ID.
func (q *Queries) UpdateTransaction(ctx context.Context, arg UpdateTransactionParams) error {
	_, err := q.exec(ctx, q.updateTransactionStmt, updateTransaction,
		arg.Amount,
		arg.Currency,
		arg.Category,
		arg.Description,
		arg.Mode,
		arg.Confirm,
		arg.TransactionDate,
		arg.ID,
	)
	return err
}
