// Package domain defines core entities, catalog definitions, and financial value types.
package domain

// Import standard library packages for timestamp representation.
import (
	// time provides representation of update and scraping timestamps.
	"time"
)

// GoldKaratPrice represents purchase and selling prices for a specific gold karat denomination.
type GoldKaratPrice struct {
	// Buy represents the consumer buy price per unit/gram in Egyptian Pounds (EGP).
	Buy float64 `json:"buy"`
	// Sell represents the consumer sell/cash-in price per unit/gram in EGP.
	Sell float64 `json:"sell"`
	// OldBuy optionally records the previous buy price for trend calculation.
	OldBuy *float64 `json:"oldBuy,omitempty"`
	// OldSell optionally records the previous sell price for trend calculation.
	OldSell *float64 `json:"oldSell,omitempty"`
	// Trend indicates whether the price moved up or down.
	Trend Trend `json:"trend"`
	// YesterdayClose optionally records the concluding price from the previous session.
	YesterdayClose *float64 `json:"yesterdayClose,omitempty"`
	// YesterdayBuyPrice optionally records the baseline prior session buy price.
	YesterdayBuyPrice float64 `json:"yesterdayBuyPrice,omitempty"`
	// YesterdaySellPrice optionally records the baseline prior session sell price.
	YesterdaySellPrice float64 `json:"yesterdaySellPrice,omitempty"`
}

// InternationalGoldQuote represents international spot ounce pricing details.
type InternationalGoldQuote struct {
	// Price specifies international spot gold price.
	Price float64 `json:"price"`
	// Buy optionally specifies spot buy price.
	Buy *float64 `json:"buy,omitempty"`
	// Sell optionally specifies spot sell price.
	Sell *float64 `json:"sell,omitempty"`
	// Currency specifies pricing denomination (typically USD).
	Currency string `json:"currency,omitempty"`
}

// GoldSnapshot represents the precious metals quote envelope returned by GET /v1/gold.
type GoldSnapshot struct {
	// Source identifies the authoritative data provider (e.g. "isagha").
	Source string `json:"source"`
	// Currency specifies the trading currency code (always "EGP" for domestic quotes).
	Currency string `json:"currency"`
	// Logo provides the public HTTPS asset logo URL for gold.
	Logo string `json:"logo"`
	// UpdatedAt records the provider's reported update timestamp.
	UpdatedAt time.Time `json:"updatedAt"`
	// ScrapedAt records when the gateway ingested the quote.
	ScrapedAt time.Time `json:"scrapedAt"`
	// Freshness indicates data reliability state ("current", "stale", "degraded").
	Freshness string `json:"freshness"`
	// SourceTraceID provides distributed request audit traceability.
	SourceTraceID string `json:"sourceTraceId"`
	// International optionally records spot ounce gold prices.
	International *InternationalGoldQuote `json:"international,omitempty"`
	// Rates maps karat denominations ("24k", "21k", "18k", "ounce", "pound") to price details.
	Rates map[string]GoldKaratPrice `json:"rates"`
}

// GasPriceItem represents fuel pump price details for an individual petroleum product.
type GasPriceItem struct {
	// Price specifies current fuel price per unit in EGP.
	Price float64 `json:"price"`
	// Unit specifies volume or packaging metric ("liter", "cylinder", "m3").
	Unit string `json:"unit"`
	// OldPrice optionally records the preceding price before the last price committee adjustment.
	OldPrice *float64 `json:"oldPrice,omitempty"`
	// Trend indicates price adjustment direction ("up", "down", or steady).
	Trend Trend `json:"trend"`
	// YesterdayClose optionally records the baseline prior session price.
	YesterdayClose *float64 `json:"yesterdayClose,omitempty"`
}

// GasSnapshot represents domestic petroleum and fuel rates returned by GET /v1/gas.
type GasSnapshot struct {
	// Source identifies the petroleum authority or data source.
	Source string `json:"source"`
	// Currency specifies the pricing currency (always "EGP").
	Currency string `json:"currency"`
	// Logo provides the public HTTPS asset logo URL for gas.
	Logo string `json:"logo"`
	// UpdatedAt records official price decree timestamp.
	UpdatedAt time.Time `json:"updatedAt"`
	// ScrapedAt records ingestion timestamp into the gateway.
	ScrapedAt time.Time `json:"scrapedAt"`
	// Freshness indicates temporal state ("current", "stale").
	Freshness string `json:"freshness"`
	// SourceTraceID provides request traceability.
	SourceTraceID string `json:"sourceTraceId"`
	// Rates maps product keys ("benzene95", "benzene92", "benzene80", "diesel", "cng") to details.
	Rates map[string]GasPriceItem `json:"rates"`
}
