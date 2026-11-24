// Package domain provides enterprise financial entities, quote models, and catalog metadata.
package domain

// QuoteItem models a single normalized asset quote across multi-asset categories.
type QuoteItem struct {
	// ID is the exact requested identifier (e.g. CBE_USD, 21k, EGX30, COMI).
	ID string `json:"id"`
	// Category is the domain category of the asset (currency, gold, index, stock).
	Category string `json:"category"`
	// Name is the canonical human-readable display name.
	Name string `json:"name"`
	// Price is the primary numeric market price or last trade value.
	Price float64 `json:"price"`
	// BuyPrice is the optional buy exchange rate.
	BuyPrice *float64 `json:"buyPrice"`
	// SellPrice is the optional sell exchange rate.
	SellPrice *float64 `json:"sellPrice"`
	// Change is the absolute price change compared to previous session.
	Change *float64 `json:"change"`
	// ChangePercent is the percentage price change compared to previous session.
	ChangePercent *float64 `json:"changePercent"`
	// Currency is the pricing denomination code (e.g. EGP, USD, SAR).
	Currency string `json:"currency"`
	// UpdatedAt is the ISO 8601 string of the latest update timestamp.
	UpdatedAt string `json:"updatedAt"`
}

// BatchQuoteRequestBody defines the JSON body structure for POST /v1/quotes.
type BatchQuoteRequestBody struct {
	// Items contains the array of asset identifier strings requested by client.
	Items []string `json:"items"`
}

// BatchQuoteResponse represents the unified response payload for batch quote queries.
type BatchQuoteResponse struct {
	// Quotes maps requested asset identifier strings to resolved QuoteItem structs.
	Quotes map[string]QuoteItem `json:"quotes"`
	// Unresolved lists requested identifier strings that could not be matched.
	Unresolved []string `json:"unresolved"`
	// Timestamp is the ISO 8601 server generation timestamp.
	Timestamp string `json:"timestamp"`
}
