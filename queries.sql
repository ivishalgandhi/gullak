-- name: CreateTransaction :many
-- Inserts a new transaction into the database.
INSERT INTO transactions (created_at, transaction_date, amount, currency, category, description, confirm)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListTransactions :many
-- Retrieves transactions optionally filtered by confirmation status and date range.
SELECT *
FROM transactions
WHERE (:confirm IS NULL OR confirm = :confirm)
  AND (:start_date IS NULL OR transaction_date >= :start_date)
  AND (:end_date IS NULL OR transaction_date <= :end_date)
ORDER BY transaction_date DESC, created_at DESC;

-- name: GetTransaction :one
-- Retrieves a single transaction by ID.
SELECT * FROM transactions WHERE id = ?;

-- name: UpdateTransaction :exec
-- Updates a transaction by ID.
UPDATE transactions
SET amount = ?, currency = ?, category = ?, description = ?, confirm = ?, transaction_date = ?
WHERE id = ?;

-- name: DeleteTransaction :exec
-- Deletes a transaction by ID.
DELETE FROM transactions WHERE id = ?;

-- name: TopExpenseCategories :many
-- Retrieves the top expense categories over a specified period.
-- Uses parameters: confirm, startDate, endDate to filter by transaction date range.
SELECT
    category,
    SUM(amount) AS total_spent
FROM transactions
WHERE transaction_date BETWEEN :startDate AND :endDate AND confirm = 1
GROUP BY category
ORDER BY total_spent DESC
LIMIT 5;-- Can be adjusted to show more or fewer categories


-- name: DailySpending :many
-- Retrieves the sum total of all transactions for each day within a specified date range.
SELECT
    transaction_date,
    SUM(amount) AS total_spent
FROM transactions
WHERE transaction_date BETWEEN :startDate AND :endDate AND confirm = 1
GROUP BY transaction_date
ORDER BY transaction_date ASC;

-- TODO: This is not live yet.
-- name: MonthlySpendingSummary :many
-- Returns the total spending grouped by month and category.
SELECT
    strftime('%Y', transaction_date) AS year,
    strftime('%m', transaction_date) AS month,
    category,
    SUM(amount) AS total_spent
FROM transactions
GROUP BY year, month, category
ORDER BY year DESC, month DESC, total_spent DESC;

-- Asset Management Queries

-- name: CreateAsset :one
-- Creates a new asset
INSERT INTO assets (
    created_at, institution_name, institution_type, 
    asset_name, current_value, currency, 
    last_updated, description, confirm
)
VALUES (
    CURRENT_TIMESTAMP, ?, ?, 
    ?, ?, ?, 
    CURRENT_TIMESTAMP, ?, ?
)
RETURNING *;

-- name: ListAssets :many
-- Lists all assets
SELECT * FROM assets 
WHERE (:confirm IS NULL OR confirm = :confirm)
ORDER BY created_at DESC;

-- name: GetAsset :one
-- Gets a specific asset by ID
SELECT * FROM assets WHERE id = ?;

-- name: GetAssetByInstitution :one
-- Gets an asset by institution name, type and asset name
SELECT * FROM assets 
WHERE institution_name = ? 
AND institution_type = ? 
AND asset_name = ?;

-- name: UpdateAssetValue :one
-- Updates an asset's value
UPDATE assets 
SET current_value = ?, 
    currency = ?,
    last_updated = CURRENT_TIMESTAMP,
    confirm = ?
WHERE id = ? 
RETURNING *;

-- name: UpdateAssetValueAndHistory :exec
-- Updates an asset's value and creates a history record
UPDATE assets 
SET current_value = ?, 
    currency = ?,
    last_updated = CURRENT_TIMESTAMP,
    confirm = ?
WHERE id = ?;

-- name: CreateAssetHistory :one
-- Creates a history record for an asset value change
INSERT INTO asset_history (
    asset_id, value_date, value, currency
)
VALUES (?, CURRENT_TIMESTAMP, ?, ?)
RETURNING *;

-- name: GetAssetHistory :many
-- Gets the value history for an asset
SELECT * FROM asset_history 
WHERE asset_id = ? 
ORDER BY value_date DESC;

-- name: DeleteAsset :exec
-- Deletes an asset
DELETE FROM assets WHERE id = ?;