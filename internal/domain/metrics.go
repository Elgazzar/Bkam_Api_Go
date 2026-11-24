// Package domain defines core entities, catalog definitions, and financial value types.
package domain

// CategoryCounts records request counts grouped by financial asset domain category.
type CategoryCounts struct {
	// Banks records aggregate request count for foreign exchange rates.
	Banks int `json:"banks"`
	// Gold records aggregate request count for precious metals commodities.
	Gold int `json:"gold"`
	// Stocks records aggregate request count for domestic and GCC equities.
	Stocks int `json:"stocks"`
	// Indices records aggregate request count for market benchmark indices.
	Indices int `json:"indices"`
	// Gas records aggregate request count for domestic fuel and petroleum.
	Gas int `json:"gas"`
	// Other records aggregate request count for discovery, catalog, and administrative queries.
	Other int `json:"other"`
}

// DailyMetrics models request metrics aggregated for a specific calendar day in Cairo timezone.
type DailyMetrics struct {
	// Date specifies the calendar day identifier (YYYY-MM-DD).
	Date string `json:"date"`
	// Total records the sum of all requests served on this calendar date.
	Total int `json:"total"`
	// Categories records the categorized distribution of requests.
	Categories CategoryCounts `json:"categories"`
}

// MetricsResponse models the JSON payload returned by GET /v1/metrics.
type MetricsResponse struct {
	// ServerID is the unique identifier for the host server process.
	ServerID string `json:"serverId"`
	// Date is the current operational calendar date in Africa/Cairo timezone.
	Date string `json:"date"`
	// Today contains the active request counts for the current calendar date.
	Today DailyMetrics `json:"today"`
	// History optionally contains historical daily records when ?days > 1 is requested.
	History []DailyMetrics `json:"history,omitempty"`
}
