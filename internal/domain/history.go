// Package domain defines core entities, catalog definitions, and financial value types.
package domain

// WeeklyBucket represents consolidated daily snapshots for an entire ISO calendar week (YYYY-Www).
type WeeklyBucket map[string]any

// ResolvedDateSnapshot represents a single calendar date archive envelope returned by historical endpoints.
type ResolvedDateSnapshot struct {
	// Data holds the concrete domain payload (rates, gold, stocks, indices).
	Data any `json:"data"`
	// IsToday indicates whether this snapshot was synthesized dynamically from the active trading session.
	IsToday bool `json:"isToday"`
	// Date specifies the ISO calendar date string (YYYY-MM-DD).
	Date string `json:"date"`
	// CacheTTLSeconds specifies recommended client HTTP Cache-Control max-age in seconds.
	CacheTTLSeconds int `json:"cacheTtlSeconds"`
}

// RangeOptions specifies parameters and filters for multi-day historical range queries.
type RangeOptions struct {
	// StartDate specifies the starting calendar date (YYYY-MM-DD).
	StartDate string `json:"startDate"`
	// EndDate specifies the concluding calendar date (YYYY-MM-DD).
	EndDate string `json:"endDate"`
	// Market optionally filters regional GCC equity exchanges ("saudi", "uae", "qatar", "kuwait").
	Market string `json:"market,omitempty"`
	// Symbol optionally filters an individual stock or index ticker symbol.
	Symbol string `json:"symbol,omitempty"`
	// Currency optionally filters foreign exchange currency or quote denomination.
	Currency string `json:"currency,omitempty"`
	// Bank optionally filters foreign exchange rates offered by a specific banking code.
	Bank string `json:"bank,omitempty"`
	// Karat optionally filters a precious metal karat denomination (e.g. "24", "21", "18").
	Karat string `json:"karat,omitempty"`
}
