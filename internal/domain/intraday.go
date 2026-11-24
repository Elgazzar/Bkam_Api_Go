// Package domain provides enterprise financial entities, quote models, and catalog metadata.
package domain

// IntradayPoint represents a single intraday price tick point in a sparkline series.
type IntradayPoint struct {
	// Time is the ISO 8601 timestamp string when the tick occurred.
	Time string `json:"time"`
	// Price is the primary trade price or exchange rate reading.
	Price float64 `json:"price"`
	// Buy is the optional buy exchange rate reading.
	Buy *float64 `json:"buy,omitempty"`
	// Sell is the optional sell exchange rate reading.
	Sell *float64 `json:"sell,omitempty"`
}

// RatesIntradayResponse defines the API response payload for GET /v1/rates/:currency/intraday/latest.
type RatesIntradayResponse struct {
	// Currency is the target currency code (e.g. USD, EUR).
	Currency string `json:"currency"`
	// Bank is the optional commercial bank identifier if filtered.
	Bank string `json:"bank,omitempty"`
	// Date is the resolved trading session date in YYYY-MM-DD format.
	Date string `json:"date"`
	// Buy is the latest buy exchange rate.
	Buy float64 `json:"buy"`
	// Sell is the latest sell exchange rate.
	Sell float64 `json:"sell"`
	// PreviousBuyPrice is the baseline previous session closing buy rate.
	PreviousBuyPrice float64 `json:"previousBuyPrice"`
	// PreviousSellPrice is the baseline previous session closing sell rate.
	PreviousSellPrice float64 `json:"previousSellPrice"`
	// UpdatedAt is the ISO 8601 update timestamp string.
	UpdatedAt string `json:"updatedAt"`
	// Intraday is the sequence of intraday points.
	Intraday []IntradayPoint `json:"intraday"`
}

// GoldKaratIntradayItem models a single karat intraday series inside the unified gold response.
type GoldKaratIntradayItem struct {
	// Buy is the current buy price.
	Buy float64 `json:"buy"`
	// Sell is the current sell price.
	Sell float64 `json:"sell"`
	// PreviousBuyPrice is the previous session baseline buy price.
	PreviousBuyPrice float64 `json:"previousBuyPrice"`
	// PreviousSellPrice is the previous session baseline sell price.
	PreviousSellPrice float64 `json:"previousSellPrice"`
	// Intraday is the sequence of intraday points for this karat.
	Intraday []IntradayPoint `json:"intraday"`
}

// GoldInternationalIntradayItem models international spot gold in the unified gold response.
type GoldInternationalIntradayItem struct {
	// Buy is the spot buy price.
	Buy float64 `json:"buy"`
	// Sell is the spot sell price.
	Sell float64 `json:"sell"`
	// PreviousBuyPrice is the baseline buy price.
	PreviousBuyPrice float64 `json:"previousBuyPrice"`
	// PreviousSellPrice is the baseline sell price.
	PreviousSellPrice float64 `json:"previousSellPrice"`
	// Currency is the denomination currency code (USD).
	Currency string `json:"currency"`
	// Intraday is the sequence of intraday points.
	Intraday []IntradayPoint `json:"intraday"`
}

// GoldIntradayResponse defines the API response payload for GET /v1/gold/intraday/latest.
type GoldIntradayResponse struct {
	// Date is the resolved trading session date.
	Date string `json:"date"`
	// Currency is the currency code (EGP).
	Currency string `json:"currency"`
	// UpdatedAt is the ISO 8601 update timestamp string.
	UpdatedAt string `json:"updatedAt"`
	// Karats contains intraday series maps for 24, 21, and 18 karat gold.
	Karats map[string]GoldKaratIntradayItem `json:"karats"`
	// Pound contains intraday series for the Egyptian Gold Pound.
	Pound GoldKaratIntradayItem `json:"pound"`
	// International contains intraday series for international spot ounce.
	International GoldInternationalIntradayItem `json:"international"`
}

// GoldSingleIntradayResponse defines the API response payload for GET /v1/gold/:karat/intraday/latest.
type GoldSingleIntradayResponse struct {
	// Item is the karat or product identifier (e.g. 21, 24, pound, international).
	Item string `json:"item"`
	// Date is the resolved trading session date.
	Date string `json:"date"`
	// Currency is the currency code.
	Currency string `json:"currency"`
	// Buy is the current buy price.
	Buy float64 `json:"buy"`
	// Sell is the current sell price.
	Sell float64 `json:"sell"`
	// PreviousBuyPrice is the baseline buy price.
	PreviousBuyPrice float64 `json:"previousBuyPrice"`
	// PreviousSellPrice is the baseline sell price.
	PreviousSellPrice float64 `json:"previousSellPrice"`
	// UpdatedAt is the ISO 8601 update timestamp string.
	UpdatedAt string `json:"updatedAt"`
	// Intraday is the sequence of intraday points.
	Intraday []IntradayPoint `json:"intraday"`
}

// StockIntradayResponse defines the API response payload for GET /v1/stocks/:symbol/intraday/latest.
type StockIntradayResponse struct {
	// Symbol is the canonical stock ticker symbol.
	Symbol string `json:"symbol"`
	// Date is the resolved trading session date.
	Date string `json:"date"`
	// Currency is the quotation currency (EGP, SAR, etc.).
	Currency string `json:"currency"`
	// Price is the last traded price.
	Price float64 `json:"price"`
	// PreviousClose is the preceding trading session close baseline price.
	PreviousClose float64 `json:"previousClose"`
	// Change is the absolute price change.
	Change float64 `json:"change"`
	// ChangePercent is the percentage price change.
	ChangePercent float64 `json:"changePercent"`
	// UpdatedAt is the ISO 8601 update timestamp string.
	UpdatedAt string `json:"updatedAt"`
	// Intraday is the sequence of intraday points.
	Intraday []IntradayPoint `json:"intraday"`
}
