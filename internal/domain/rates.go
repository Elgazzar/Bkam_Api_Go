// Package domain defines core entities, catalog definitions, and financial value types.
package domain

// Import standard library packages for timestamp representation and formatting.
import (
	// time provides representation of retrieval and bank updated timestamps.
	"time"
)

// Trend represents directional price movement relative to prior quotes ('up' or 'down').
type Trend string

// Predefined Trend constants.
const (
	// TrendUp indicates increasing price movement.
	TrendUp Trend = "up"
	// TrendDown indicates decreasing price movement.
	TrendDown Trend = "down"
)

// CurrencyRate represents an enriched foreign exchange rate quote offered by a specific bank.
type CurrencyRate struct {
	// NameAr represents the localized Arabic name of the financial institution.
	NameAr string `json:"name_ar"`
	// NameEn represents the localized English name of the financial institution.
	NameEn string `json:"name_en"`
	// Logo represents the HTTPS URL to the bank's brand logo image.
	Logo string `json:"logo"`
	// BankCode represents the unique uppercase symbol identifying the bank (e.g. "NBE").
	BankCode string `json:"bankCode"`
	// CurrencyCode represents the 3-letter ISO 4217 uppercase currency code (e.g. "USD").
	CurrencyCode string `json:"currencyCode"`
	// BuyRate represents the bank's purchase exchange rate.
	BuyRate float64 `json:"buyRate"`
	// SellRate represents the bank's selling exchange rate.
	SellRate float64 `json:"sellRate"`
	// Trend indicates whether the rate increased or decreased ('up' or 'down').
	Trend Trend `json:"trend"`
	// OldBuyRate optionally records the prior recorded buy rate for comparison.
	OldBuyRate *float64 `json:"oldBuyRate,omitempty"`
	// OldSellRate optionally records the prior recorded sell rate for comparison.
	OldSellRate *float64 `json:"oldSellRate,omitempty"`
	// RetrievedAt records the exact timestamp when this rate was scraped and stored.
	RetrievedAt time.Time `json:"retrievedAt"`
	// BankUpdatedAt optionally records when the bank's official portal updated the rate.
	BankUpdatedAt *time.Time `json:"bankUpdatedAt,omitempty"`
	// Freshness indicates temporal state ("current", "stale", "degraded").
	Freshness string `json:"freshness"`
	// SourceTraceID provides distributed request traceability for upstream audits.
	SourceTraceID string `json:"sourceTraceId"`
	// DailyHighBuy optionally records the highest recorded purchase rate during the current day.
	DailyHighBuy *float64 `json:"dailyHighBuy,omitempty"`
	// DailyHighSell optionally records the highest recorded selling rate during the current day.
	DailyHighSell *float64 `json:"dailyHighSell,omitempty"`
	// DailyLowBuy optionally records the lowest recorded purchase rate during the current day.
	DailyLowBuy *float64 `json:"dailyLowBuy,omitempty"`
	// DailyLowSell optionally records the lowest recorded selling rate during the current day.
	DailyLowSell *float64 `json:"dailyLowSell,omitempty"`
	// YesterdayCloseBuy optionally records the concluding purchase rate of the prior trading session.
	YesterdayCloseBuy *float64 `json:"yesterdayCloseBuy,omitempty"`
	// YesterdayCloseSell optionally records the concluding selling rate of the prior trading session.
	YesterdayCloseSell *float64 `json:"yesterdayCloseSell,omitempty"`
	// LastCloseDate optionally indicates the date string of the prior session close.
	LastCloseDate *string `json:"lastCloseSnapshotDate,omitempty"`
}

// AvailabilityRecord reflects the operational health and scraping state for an individual bank.
type AvailabilityRecord struct {
	// BankCode is the unique uppercase code of the bank.
	BankCode string `json:"bankCode"`
	// Status indicates availability health ("current", "stale", "outage").
	Status string `json:"status"`
	// LastSuccessAt records the timestamp of the last successful scrape.
	LastSuccessAt time.Time `json:"lastSuccessAt"`
}

// RateSnapshot represents the public exchange rates snapshot envelope returned by GET /v1/rates.
type RateSnapshot struct {
	// RequestedAt records when the snapshot query was processed by the gateway.
	RequestedAt time.Time `json:"requestedAt"`
	// Freshness indicates consolidated health ("current", "partial", "stale").
	Freshness string `json:"freshness"`
	// RequestFilters mirrors query filters applied to this response (e.g. {"currency": "USD"}).
	RequestFilters map[string]string `json:"requestFilters"`
	// Rates contains the array of individual bank exchange rate quotes.
	Rates []CurrencyRate `json:"rates"`
	// Availability summarizes scraping health across all monitored banks.
	Availability []AvailabilityRecord `json:"availability"`
	// Warnings optionally communicates non-fatal degraded conditions or provider outages.
	Warnings []string `json:"warnings"`
	// Currency optionally records the single currency filter if requested.
	Currency string `json:"currency,omitempty"`
}
