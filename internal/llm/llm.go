package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/mr-karan/gullak/pkg/models"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type Manager struct {
	log    *slog.Logger
	client *openai.Client
	model  string
}

func New(token, baseURL, model string, timeout time.Duration, log *slog.Logger) (*Manager, error) {
	// Initialize the OpenAI client.
	cfg := openai.DefaultConfig(token)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	if timeout > 0 {
		cfg.HTTPClient.Timeout = timeout
	} else {
		// Set a default timeout of 10 seconds.
		cfg.HTTPClient.Timeout = 10 * time.Second
	}

	client := openai.NewClientWithConfig(cfg)

	return &Manager{
		client: client,
		model:  model,
		log:    log,
	}, nil
}

// Combined struct for parsing both transactions and assets
type FinancialData struct {
	Transactions []models.Item `json:"transactions,omitempty"`
	Assets       []models.Asset `json:"assets,omitempty"`
}

// Custom error types
type NoValidTransactionError struct {
	Message string
}

func (e *NoValidTransactionError) Error() string {
	return e.Message
}

type NoValidAssetError struct {
	Message string
}

func (e *NoValidAssetError) Error() string {
	return e.Message
}

// Parse combines expense and asset parsing into a single method
func (m *Manager) Parse(msg string) (FinancialData, error) {
	if msg == "" {
		return FinancialData{}, errors.New("empty message")
	}

	m.log.Debug("Parsing financial data", "message", msg)

	// Define a comprehensive function for parsing both expenses and assets
	fnParseFinancialData := openai.FunctionDefinition{
		Name:        "parse_financial_data",
		Description: "Parse financial data including expenses and assets from natural language input.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				// Transactions schema (similar to previous implementation)
				"transactions": {
					Type:        jsonschema.Array,
					Description: "List of expenses or transactions",
					Items: &jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"transaction_date": {
								Type:        jsonschema.String,
								Description: "Date of transaction in ISO 8601 format (e.g., 2021-09-01) if specified else today's date.",
							},
							"amount": {
								Type:        jsonschema.Number,
								Description: "Amount of the transaction",
							},
							"category": {
								Type:        jsonschema.String,
								Description: "One word category of the expense (e.g., food, travel, entertainment)",
							},
							"description": {
								Type:        jsonschema.String,
								Description: "Concise and short description of the item",
							},
						},
						Required: []string{"transaction_date", "amount", "category", "description"},
					},
				},
				// Assets schema (similar to previous implementation)
				"assets": {
					Type:        jsonschema.Array,
					Description: "List of assets",
					Items: &jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"institution_name": {
								Type:        jsonschema.String,
								Description: "Name of the institution (e.g., HDFC Bank, Zerodha)",
							},
							"institution_type": {
								Type:        jsonschema.String,
								Description: "Type of institution (e.g., bank, broker, mutual_fund, other)",
								Enum:        []string{"bank", "broker", "mutual_fund", "other"},
							},
							"asset_name": {
								Type:        jsonschema.String,
								Description: "Name of the asset (e.g., Savings Account, Stock Portfolio)",
							},
							"current_value": {
								Type:        jsonschema.Number,
								Description: "Current value of the asset",
							},
							"currency": {
								Type:        jsonschema.String,
								Description: "Currency of the asset value (default: USD)",
							},
							"description": {
								Type:        jsonschema.String,
								Description: "Additional description of the asset",
							},
							"confirm": {
								Type:        jsonschema.Boolean,
								Description: "Whether the asset entry is confirmed",
							},
						},
						Required: []string{"institution_name", "institution_type", "asset_name", "current_value"},
					},
				},
			},
		},
	}

	// Prepare the dialogue with a comprehensive system prompt
	dialogue := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem, 
			Content: fmt.Sprintf(`You are a financial assistant that helps parse financial data from natural language. 
Your task is to extract:
1. Expenses/Transactions: Categorize spending in valid categories
2. Assets: Capture information about financial assets

For assets, follow these rules:
- Institution Name: Name of the bank/broker (e.g., "HDFC", "Citibank", "Zerodha")
- Institution Type: One of: "bank", "broker", "mutual_fund", "other"
- Asset Name: Should be concise and NOT include the institution name. Examples:
  "Fixed Deposit" (not "Citibank Fixed Deposit")
  "Savings Account" (not "HDFC Savings Account")
  "Stock Portfolio or Portffolio" (not "Zerodha Portfolio")

Example inputs and expected parsing:
- Input: "I have $5000 in my HDFC savings account"
  Output: {institution: "HDFC", type: "bank", asset: "Savings Account"}
- Input: "My Zerodha portfolio is worth ₹100000"
  Output: {institution: "Zerodha", type: "broker", asset: "Stock Portfolio"}
- Input: "Added 10000 to my HDFC mutual fund"
  Output: {institution: "HDFC", type: "mutual_fund", asset: "Mutual Fund"}

Today's date is %s`, time.Now().Format("2006-01-02")),
		},
		{Role: openai.ChatMessageRoleUser, Content: msg},
	}

	// Prepare the tool for financial data parsing
	t := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &fnParseFinancialData,
	}

	// Create chat completion
	resp, err := m.client.CreateChatCompletion(context.TODO(),
		openai.ChatCompletionRequest{
			Model:    m.model,
			Messages: dialogue,
			Tools:    []openai.Tool{t},
		},
	)

	if err != nil || len(resp.Choices) != 1 {
		m.log.Error("Completion error", "error", err, "choices", len(resp.Choices))
		return FinancialData{}, fmt.Errorf("error completing the request")
	}

	var financialData FinancialData

	for _, choice := range resp.Choices {
		for _, toolCall := range choice.Message.ToolCalls {
			if toolCall.Function.Name == fnParseFinancialData.Name {
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &financialData); err != nil {
					return FinancialData{}, fmt.Errorf("error unmarshalling response: %s", err)
				}
				
				// Set default currency for assets if not provided
				for i := range financialData.Assets {
					if financialData.Assets[i].Currency == "" {
						financialData.Assets[i].Currency = "USD"
					}
				}
				
				return financialData, nil
			}
		}
	}

	// Fallback error handling
	if len(financialData.Transactions) == 0 && len(financialData.Assets) == 0 {
		for _, choice := range resp.Choices {
			if choice.FinishReason == "stop" {
				return FinancialData{}, &NoValidTransactionError{Message: choice.Message.Content}
			}
		}
	}

	return FinancialData{}, fmt.Errorf("no valid financial data found in response")
}

// Deprecated methods kept for backward compatibility
func (m *Manager) ParseAsset(msg string) (models.Assets, error) {
	financialData, err := m.Parse(msg)
	if err != nil {
		return models.Assets{}, err
	}
	
	return models.Assets{
		Assets: financialData.Assets,
	}, nil
}

func (m *Manager) ParseTransactions(msg string) (models.Transactions, error) {
	financialData, err := m.Parse(msg)
	if err != nil {
		return models.Transactions{}, err
	}
	
	return models.Transactions{
		Transactions: financialData.Transactions,
	}, nil
}