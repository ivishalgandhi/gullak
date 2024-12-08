package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mr-karan/gullak/internal/db"
	"github.com/mr-karan/gullak/pkg/models"
)

type ExpenseInput struct {
	Line string `json:"line"`
}

type Resp struct {
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data"`
}

type CategorySummary struct {
	Category   string  `json:"category"`
	TotalSpent float64 `json:"total_spent"`
}

type DailySpendingSummary struct {
	TransactionDate string  `json:"transaction_date"`
	TotalSpent      float64 `json:"total_spent"`
}

func handleIndex(c echo.Context) error {
	return c.JSON(http.StatusOK, Resp{
		Message: "Welcome to Gullak. POST to /api/transactions to save expenses.",
	})
}

func handleCreateTransaction(c echo.Context) error {
	app := c.Get("app").(*App)

	// Parse input
	var input ExpenseInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format")
	}

	// Parse the input using unified Parse method
	financialData, err := app.llm.Parse(input.Line)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Could not parse input as transaction or asset")
	}

	// Handle assets if present
	if len(financialData.Assets) > 0 {
		for _, asset := range financialData.Assets {
			// Check if asset already exists
			existingAsset, err := app.queries.GetAssetByInstitution(c.Request().Context(), db.GetAssetByInstitutionParams{
				InstitutionName: asset.InstitutionName,
				InstitutionType: asset.InstitutionType,
				AssetName:      asset.AssetName,
			})
			
			if err == nil {
				// Asset exists, update its value and create history
				err = app.queries.UpdateAssetValueAndHistory(c.Request().Context(), db.UpdateAssetValueAndHistoryParams{
					CurrentValue: asset.CurrentValue,
					Currency:    asset.Currency,
					Confirm:     asset.Confirm,
					ID:          existingAsset.ID,
				})
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update asset")
				}

				// Create history record
				_, err = app.queries.CreateAssetHistory(c.Request().Context(), db.CreateAssetHistoryParams{
					AssetID:  existingAsset.ID,
					Value:    asset.CurrentValue,
					Currency: asset.Currency,
				})
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create asset history")
				}

				// Return the updated asset
				updatedAsset, err := app.queries.GetAsset(c.Request().Context(), existingAsset.ID)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch updated asset")
				}
				return c.JSON(http.StatusOK, updatedAsset)
			}

			// Asset doesn't exist, create new one
			result, err := app.queries.CreateAsset(c.Request().Context(), db.CreateAssetParams{
				InstitutionName: asset.InstitutionName,
				InstitutionType: asset.InstitutionType,
				AssetName:       asset.AssetName,
				CurrentValue:    asset.CurrentValue,
				Currency:        asset.Currency,
				Description:     asset.Description,
				Confirm:        asset.Confirm,
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create asset")
			}

			// Create initial history record for new asset
			_, err = app.queries.CreateAssetHistory(c.Request().Context(), db.CreateAssetHistoryParams{
				AssetID:  result.ID,
				Value:    asset.CurrentValue,
				Currency: asset.Currency,
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create asset history")
			}

			return c.JSON(http.StatusCreated, result)
		}
	}

	// Handle transactions if present
	if len(financialData.Transactions) > 0 {
		var createdTransactions []interface{}
		for _, t := range financialData.Transactions {
			// Parse the transaction date string into time.Time
			transactionDate, err := time.Parse("2006-01-02", t.TransactionDate)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid transaction date format. Expected YYYY-MM-DD")
			}

			result, err := app.queries.CreateTransaction(c.Request().Context(), db.CreateTransactionParams{
				TransactionDate: transactionDate,
				Amount:         t.Amount,
				Currency:       t.Currency,
				Category:       t.Category,
				Description:    t.Description,
				Confirm:        t.Confirm,
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create transaction")
			}
			createdTransactions = append(createdTransactions, result)
		}
		return c.JSON(http.StatusCreated, createdTransactions)
	}

	return echo.NewHTTPError(http.StatusBadRequest, "No valid financial data found in input")
}

func handleListTransactions(c echo.Context) error {
	m := c.Get("app").(*App)

	var params db.ListTransactionsParams

	if confirmStr := c.QueryParam("confirm"); confirmStr != "" {
		// Convert and check the confirm parameter
		confirm, err := strconv.ParseBool(confirmStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Resp{Error: "Invalid confirm value"})
		}
		params.Confirm = confirm
	} else {
		params.Confirm = nil // Explicitly setting as nil if not provided
	}

	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Resp{Error: "Invalid start date format"})
		}
		params.StartDate = startDate
	} else {
		params.StartDate = nil // Explicitly setting as nil if not provided
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Resp{Error: "Invalid end date format"})
		}
		params.EndDate = endDate
	} else {
		params.EndDate = nil // Explicitly setting as nil if not provided
	}

	// Validate the date range if both dates are provided
	if startDateStr != "" && endDateStr != "" {
		if err := validateDateRange(startDate, endDate); err != nil {
			return c.JSON(http.StatusBadRequest, Resp{
				Error: err.Error(),
			})
		}
	}

	transactions, err := m.queries.ListTransactions(context.Background(), params)
	if err != nil {
		m.log.Error("Error retrieving transactions", "error", err)
		return c.JSON(http.StatusInternalServerError, Resp{Error: "Error retrieving transactions"})
	}

	return c.JSON(http.StatusOK, Resp{
		Data:    transactions,
		Message: "Transactions retrieved",
	})
}

func handleGetTransaction(c echo.Context) error {
	m := c.Get("app").(*App)
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		m.log.Error("Invalid transaction ID", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid transaction ID",
		})
	}

	transaction, err := m.queries.GetTransaction(context.Background(), id)
	if err != nil {
		m.log.Error("Error retrieving transaction", "error", err)
		return c.JSON(http.StatusNotFound, Resp{
			Error: "Transaction not found",
		})
	}

	return c.JSON(http.StatusOK, Resp{
		Data:    transaction,
		Message: "Transaction retrieved",
	})
}

func handleUpdateTransaction(c echo.Context) error {
	m := c.Get("app").(*App)
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		m.log.Error("Invalid transaction ID", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid transaction ID",
		})
	}

	var input models.Item // Ensure models.Item has the appropriate fields
	if err := c.Bind(&input); err != nil {
		m.log.Error("Error binding input", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid input",
		})
	}

	// Ensure transaction_date is in the correct format
	transactionDate, err := time.Parse("2006-01-02", input.TransactionDate)
	if err != nil {
		m.log.Error("Error parsing transaction date", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid transaction date",
		})
	}

	params := db.UpdateTransactionParams{
		Amount:          input.Amount,
		Currency:        input.Currency,
		Category:        input.Category,
		Description:     input.Description,
		Confirm:         input.Confirm,
		TransactionDate: transactionDate,
		ID:              id,
	}

	if err := m.queries.UpdateTransaction(context.Background(), params); err != nil {
		m.log.Error("Error updating transaction", "error", err)
		return c.JSON(http.StatusInternalServerError, Resp{
			Error: "Error updating transaction",
		})
	}

	return c.JSON(http.StatusOK, Resp{
		Message: "Transaction updated",
		Data:    params,
	})
}

func handleDeleteTransaction(c echo.Context) error {
	m := c.Get("app").(*App)
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		m.log.Error("Invalid transaction ID", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid transaction ID",
		})
	}

	if err := m.queries.DeleteTransaction(context.Background(), id); err != nil {
		m.log.Error("Error deleting transaction", "error", err)
		return c.JSON(http.StatusInternalServerError, Resp{
			Error: "Error deleting transaction",
		})
	}

	return c.JSON(http.StatusOK, Resp{
		Message: "Transaction deleted",
	})
}

func handleTopExpenseCategories(c echo.Context) error {
	m := c.Get("app").(*App)
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	if startDateStr == "" || endDateStr == "" {
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Missing required parameters: start_date, end_date",
		})
	}

	// Parse start date and end date strings into time.Time
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		m.log.Error("Invalid start date", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid start date format, use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		m.log.Error("Invalid end date", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid end date format, use YYYY-MM-DD",
		})
	}

	// Validate the date range
	if err := validateDateRange(startDate, endDate); err != nil {
		return c.JSON(http.StatusBadRequest, Resp{
			Error: err.Error(),
		})
	}

	params := db.TopExpenseCategoriesParams{
		StartDate: startDate,
		EndDate:   endDate,
	}

	rawCategories, err := m.queries.TopExpenseCategories(context.Background(), params)
	if err != nil {
		m.log.Error("Error retrieving top expense categories", "error", err)
		return c.JSON(http.StatusInternalServerError, Resp{
			Error: "Error retrieving top expense categories",
		})
	}

	// Transform into client-friendly structure
	categories := make([]CategorySummary, len(rawCategories))
	for i, cat := range rawCategories {
		totalSpent := 0.0
		if cat.TotalSpent.Valid {
			totalSpent = cat.TotalSpent.Float64
		}
		categories[i] = CategorySummary{
			Category:   cat.Category,
			TotalSpent: totalSpent,
		}
	}

	return c.JSON(http.StatusOK, Resp{
		Data:    categories,
		Message: "Top expense categories retrieved",
	})
}

func handleDailySpending(c echo.Context) error {
	m := c.Get("app").(*App)
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	if startDateStr == "" || endDateStr == "" {
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Missing required parameters: start_date, end_date",
		})
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		m.log.Error("Invalid start date", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid start date format, use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		m.log.Error("Invalid end date", "error", err)
		return c.JSON(http.StatusBadRequest, Resp{
			Error: "Invalid end date format, use YYYY-MM-DD",
		})
	}

	// Validate the date range
	if err := validateDateRange(startDate, endDate); err != nil {
		return c.JSON(http.StatusBadRequest, Resp{
			Error: err.Error(),
		})
	}

	params := db.DailySpendingParams{
		StartDate: startDate,
		EndDate:   endDate,
	}

	rawSpending, err := m.queries.DailySpending(context.Background(), params)
	if err != nil {
		m.log.Error("Error retrieving daily spending", "error", err)
		return c.JSON(http.StatusInternalServerError, Resp{
			Error: "Error retrieving daily spending",
		})
	}

	// Transform into client-friendly structure
	spendingSummaries := make([]DailySpendingSummary, len(rawSpending))
	for i, daily := range rawSpending {
		totalSpent := 0.0
		if daily.TotalSpent.Valid {
			totalSpent = daily.TotalSpent.Float64
		}
		spendingSummaries[i] = DailySpendingSummary{
			TransactionDate: daily.TransactionDate.Format("2006-01-02"),
			TotalSpent:      totalSpent,
		}
	}

	return c.JSON(http.StatusOK, Resp{
		Data:    spendingSummaries,
		Message: "Daily spending totals retrieved successfully",
	})
}

// Asset Management Handlers

func handleCreateAsset(c echo.Context) error {
	app := c.Get("app").(*App)
	var asset models.Asset
	if err := c.Bind(&asset); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Create asset in database
	result, err := app.queries.CreateAsset(c.Request().Context(), db.CreateAssetParams{
		InstitutionName: asset.InstitutionName,
		InstitutionType: asset.InstitutionType,
		AssetName:       asset.AssetName,
		CurrentValue:    asset.CurrentValue,
		Currency:        asset.Currency,
		Description:     asset.Description,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create asset")
	}

	return c.JSON(http.StatusCreated, result)
}

func handleListAssets(c echo.Context) error {
	app := c.Get("app").(*App)
	assets, err := app.queries.ListAssets(c.Request().Context(), nil) // Pass nil to get all assets regardless of confirmation status
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch assets")
	}
	return c.JSON(http.StatusOK, assets)
}

func handleGetAsset(c echo.Context) error {
	app := c.Get("app").(*App)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid asset ID")
	}
	
	asset, err := app.queries.GetAsset(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Asset not found")
	}
	return c.JSON(http.StatusOK, asset)
}

func handleUpdateAssetValue(c echo.Context) error {
	app := c.Get("app").(*App)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid asset ID")
	}

	var update struct {
		CurrentValue float64 `json:"current_value"`
		Currency    string  `json:"currency"`
		Confirm     bool    `json:"confirm"`
	}
	if err := c.Bind(&update); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Update asset value
	result, err := app.queries.UpdateAssetValue(c.Request().Context(), db.UpdateAssetValueParams{
		ID:           id,
		CurrentValue: update.CurrentValue,
		Currency:     update.Currency,
		Confirm:      update.Confirm,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update asset value")
	}

	// Create history entry
	_, err = app.queries.CreateAssetHistory(c.Request().Context(), db.CreateAssetHistoryParams{
		AssetID:  id,
		Value:    update.CurrentValue,
		Currency: update.Currency,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create asset history")
	}

	return c.JSON(http.StatusOK, result)
}

func handleDeleteAsset(c echo.Context) error {
	app := c.Get("app").(*App)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid asset ID")
	}

	err = app.queries.DeleteAsset(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete asset")
	}
	return c.NoContent(http.StatusOK)
}

func handleGetAssetHistory(c echo.Context) error {
	app := c.Get("app").(*App)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid asset ID")
	}

	history, err := app.queries.GetAssetHistory(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch asset history")
	}
	return c.JSON(http.StatusOK, history)
}

// validateDateRange ensures that the start date is before or the same as the end date.
func validateDateRange(startDate, endDate time.Time) error {
	if startDate.After(endDate) {
		return errors.New("start date must be on or before end date")
	}
	return nil
}
