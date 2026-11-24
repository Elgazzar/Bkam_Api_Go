// Package domain defines core entities, catalog definitions, and financial value types.
package domain

// Import standard library packages for timestamp representation.
import (
	// time represents quote update and retrieval timestamps.
	"time"
)

// StockPriceItem represents real-time trading statistics and quote details for an equity security.
type StockPriceItem struct {
	// Symbol is the standard ticker symbol (e.g. "COMI", "1010").
	Symbol string `json:"symbol"`
	// Code optionally provides the numerical or exchange symbol code.
	Code string `json:"code,omitempty"`
	// NumericCode optionally records the numeric exchange code for regional equities.
	NumericCode string `json:"numericCode,omitempty"`
	// NameAr optionally provides the localized Arabic name of the listed enterprise.
	NameAr string `json:"name_ar,omitempty"`
	// NameEn optionally provides the localized English name of the listed enterprise.
	NameEn string `json:"name_en,omitempty"`
	// Price specifies the latest execution or closing trade price.
	Price float64 `json:"price"`
	// Change specifies absolute price difference from prior session close.
	Change float64 `json:"change"`
	// ChangePercent specifies percentage price difference from prior session close.
	ChangePercent float64 `json:"changePercent"`
	// Currency optionally records equity quotation currency denomination.
	Currency string `json:"currency,omitempty"`
	// Volume optionally records cumulative traded share volume during the session.
	Volume int64 `json:"volume,omitempty"`
	// Value optionally records cumulative traded monetary value during the session.
	Value float64 `json:"value,omitempty"`
	// Open optionally records the opening auction price.
	Open *float64 `json:"open,omitempty"`
	// High optionally records the peak traded price during the active session.
	High *float64 `json:"high,omitempty"`
	// Low optionally records the lowest traded price during the active session.
	Low *float64 `json:"low,omitempty"`
	// Close optionally records the official session closing price.
	Close *float64 `json:"close,omitempty"`
	// PreviousClose optionally records the baseline prior session price.
	PreviousClose float64 `json:"previousClose,omitempty"`
	// YesterdayClose optionally records the concluding price from the previous trading day.
	YesterdayClose *float64 `json:"yesterdayClose,omitempty"`
	// Trend indicates whether the price gained or lost value ('up' or 'down').
	Trend Trend `json:"trend"`
	// TradeCount optionally records the total number of executed transactions.
	TradeCount int64 `json:"tradeCount,omitempty"`
}

// StockSnapshot represents equities for a specific stock exchange (EGX, Saudi, UAE, Qatar, Kuwait).
type StockSnapshot struct {
	// Source identifies data feed provider (e.g. "egx", "tadawul", "mubasher").
	Source string `json:"source"`
	// Market identifies regional exchange code ("egx", "saudi", "uae", "qatar", "kuwait").
	Market string `json:"market"`
	// Currency specifies local trading currency ("EGP", "SAR", "AED", "QAR", "KWD").
	Currency string `json:"currency"`
	// UpdatedAt records provider exchange timestamp.
	UpdatedAt time.Time `json:"updatedAt"`
	// ScrapedAt records ingestion timestamp into the gateway.
	ScrapedAt time.Time `json:"scrapedAt"`
	// Freshness indicates temporal state ("current", "stale").
	Freshness string `json:"freshness"`
	// SourceTraceID provides distributed request audit traceability.
	SourceTraceID string `json:"sourceTraceId"`
	// Stocks maps uppercase equity symbols to StockPriceItem quote details.
	Stocks map[string]StockPriceItem `json:"stocks"`
}
