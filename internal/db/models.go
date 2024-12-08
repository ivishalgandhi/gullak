// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"time"
)

type Asset struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	InstitutionName string    `json:"institution_name"`
	InstitutionType string    `json:"institution_type"`
	AssetName       string    `json:"asset_name"`
	CurrentValue    float64   `json:"current_value"`
	Currency        string    `json:"currency"`
	LastUpdated     time.Time `json:"last_updated"`
	Description     string    `json:"description"`
	Confirm         bool      `json:"confirm"`
}

type AssetHistory struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	ValueDate time.Time `json:"value_date"`
	Value     float64   `json:"value"`
	Currency  string    `json:"currency"`
}

type Transaction struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	TransactionDate time.Time `json:"transaction_date"`
	Currency        string    `json:"currency"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	Description     string    `json:"description"`
	Confirm         bool      `json:"confirm"`
}
