// Package domain defines core entities, catalog definitions, and financial value types.
package domain

// Import standard library packages for timestamp representation.
import (
	// time represents index quote update and scraping timestamps.
	"time"
)

// IndexPriceItem represents market benchmark index figures and percentage movements.
type IndexPriceItem struct {
	// Code specifies the index code symbol required by indexItemSchema (e.g. "EGX30", "TASI").
	Code string `json:"code"`
	// IndexCode is the canonical benchmark code (e.g. "EGX30", "EGX70", "TASI").
	IndexCode string `json:"indexCode,omitempty"`
	// Name specifies the human-readable index title.
	Name string `json:"name,omitempty"`
	// NameEn optionally specifies the English index title.
	NameEn string `json:"nameEn,omitempty"`
	// NameAr optionally specifies the localized Arabic index title.
	NameAr string `json:"nameAr,omitempty"`
	// Price specifies current points value of the benchmark index.
	Price float64 `json:"price"`
	// Change specifies absolute point difference relative to prior session close.
	Change float64 `json:"change"`
	// ChangePercent specifies percentage point difference relative to prior session close.
	ChangePercent float64 `json:"changePercent"`
	// High optionally records session high points value.
	High *float64 `json:"high,omitempty"`
	// Low optionally records session low points value.
	Low *float64 `json:"low,omitempty"`
	// YesterdayClose optionally records concluding points value from the previous session.
	YesterdayClose *float64 `json:"yesterdayClose,omitempty"`
	// Trend indicates point movement direction ('up' or 'down').
	Trend Trend `json:"trend"`
}

// IndexSnapshot represents index benchmarks for an exchange returned by GET /v1/indices.
type IndexSnapshot struct {
	// Source identifies data feed provider.
	Source string `json:"source"`
	// Market identifies regional exchange code ("egx", "saudi", "uae", "qatar", "kuwait").
	Market string `json:"market"`
	// Currency specifies pricing currency denomination.
	Currency string `json:"currency"`
	// UpdatedAt records provider exchange timestamp.
	UpdatedAt time.Time `json:"updatedAt"`
	// ScrapedAt records gateway ingestion timestamp.
	ScrapedAt time.Time `json:"scrapedAt"`
	// Freshness indicates temporal state ("current", "stale").
	Freshness string `json:"freshness"`
	// SourceTraceID provides distributed audit traceability.
	SourceTraceID string `json:"sourceTraceId"`
	// Indices maps uppercase index codes (e.g. "EGX30") to IndexPriceItem details.
	Indices map[string]IndexPriceItem `json:"indices"`
}
