package models

type Item struct {
	ID              int64   `json:"id"`
	CreatedAt       string  `json:"created_at"`
	TransactionDate string  `json:"transaction_date"`
	Currency        string  `json:"currency"`
	Amount          float64 `json:"amount"`
	Category        string  `json:"category"`
	Description     string  `json:"description"`
	Confirm         bool    `json:"confirm"`
}

type Transactions struct {
	Transactions []Item `json:"transactions"`
}

type Asset struct {
    InstitutionName string  `json:"institution_name"`
    InstitutionType string  `json:"institution_type"`
    AssetName       string  `json:"asset_name"`
    CurrentValue    float64 `json:"current_value"`
    Currency        string  `json:"currency,omitempty"`
    Description     string  `json:"description,omitempty"`
    Confirm         bool    `json:"confirm"`
}

type Assets struct {
    Assets []Asset `json:"assets"`
}
