// Package domain defines core entities, catalog definitions, and financial value types.
package domain

// Import standard library packages for embedded filesystem and JSON handling.
import (
	// embed enables embedding regional stock catalog JSON directly into the binary.
	_ "embed"
	// encoding/json parses static catalog JSON into typed structs.
	"encoding/json"
	// sync provides atomic once initialization.
	"sync"
)

//go:embed regional_catalog.json
var rawRegionalCatalogData []byte

// RegionalMarketDefinition describes a GCC regional stock market exchange.
type RegionalMarketDefinition struct {
	// Country is the ISO country code (e.g. "SA", "AE", "QA", "KW").
	Country string `json:"country"`
	// Name is the market title.
	Name string `json:"name"`
	// Exchange is the primary exchange code ("TADAWUL", "DFM", "QSE", "KSE").
	Exchange string `json:"exchange"`
	// Currency is local trading currency ("SAR", "AED", "QAR", "KWD").
	Currency string `json:"currency"`
	// Timezone is IANA timezone string.
	Timezone string `json:"timezone"`
	// TradingDays lists active session days.
	TradingDays []string `json:"tradingDays"`
	// TradingHours describes local session hours.
	TradingHours string `json:"tradingHours"`
	// StocksCount counts listed securities.
	StocksCount int `json:"stocksCount"`
	// Stocks contains full catalog array.
	Stocks []any `json:"stocks"`
}

// RegionalCatalogPayload contains all regional GCC stock market definitions.
type RegionalCatalogPayload struct {
	// Markets contains definitions for SA, AE, QA, and KW.
	Markets []RegionalMarketDefinition `json:"markets"`
}

// globalRegionalCatalog holds the cached singleton payload.
var (
	// globalRegionalCatalog is the populated pointer.
	globalRegionalCatalog *RegionalCatalogPayload
	// regionalCatalogOnce ensures atomic single initialization.
	regionalCatalogOnce sync.Once
)

// GetRegionalCatalog returns the cached RegionalCatalogPayload.
func GetRegionalCatalog() *RegionalCatalogPayload {
	// Ensure one-time initialization.
	regionalCatalogOnce.Do(func() {
		// Initialize empty payload struct.
		var payload RegionalCatalogPayload
		// Unmarshal embedded JSON data.
		if err := json.Unmarshal(rawRegionalCatalogData, &payload); err == nil {
			// Assign global pointer.
			globalRegionalCatalog = &payload
		}
	})
	// Return cached pointer.
	return globalRegionalCatalog
}
