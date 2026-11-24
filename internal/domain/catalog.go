// Package domain defines core entities, catalog definitions, and financial value types.
package domain

// Import standard library packages for embedded file system, JSON parsing, and string utilities.
import (
	// embed enables compiling static assets directly into the Go executable binary.
	_ "embed"
	// encoding/json parses static JSON catalog payloads into typed Go structures.
	"encoding/json"
	// strings provides uppercase normalization and whitespace trimming for lookups.
	"strings"
	// sync provides thread-safe initialization primitives.
	"sync"
)

//go:embed catalog_data.json
var rawCatalogData []byte

// Bank represents an institutional bank entity operating foreign exchange rate feeds in Egypt.
type Bank struct {
	// Code represents the unique uppercase symbol identifying the bank (e.g. "NBE", "CIB").
	Code string `json:"code"`
	// Name represents the human-readable default name of the bank in English.
	Name string `json:"name"`
	// NameAr represents the localized Arabic name of the financial institution.
	NameAr string `json:"name_ar"`
	// NameEn represents the localized English name of the financial institution.
	NameEn string `json:"name_en"`
	// Logo represents the absolute HTTPS URL pointing to the bank's brand logo image.
	Logo string `json:"logo"`
	// SourceURL points to the public foreign exchange rates portal scraped by background daemons.
	SourceURL string `json:"sourceUrl"`
	// Enabled indicates whether the bank's exchange rate feeds are active.
	Enabled bool `json:"enabled"`
	// Availability indicates current operational status ("current", "stale", "degraded").
	Availability string `json:"availability"`
}

// Currency represents a fiat currency recognized and converted by Egyptian banks.
type Currency struct {
	// Code represents the 3-letter ISO 4217 uppercase currency code (e.g. "USD", "EUR").
	Code string `json:"code"`
	// Name represents the international human-readable title of the currency.
	Name string `json:"name"`
	// MinorUnit represents decimal precision for the currency (typically 4 for FX rates).
	MinorUnit int `json:"minorUnit"`
}

// CommodityDefinition represents a precious metal or fuel product.
type CommodityDefinition struct {
	// Code represents the unique identifier for the commodity (e.g. "24", "21", "BENZENE95").
	Code string `json:"code"`
	// Category indicates the commodity class ("gold" or "fuel").
	Category string `json:"category"`
	// NameAr represents the localized Arabic title.
	NameAr string `json:"nameAr"`
	// NameEn represents the localized English title.
	NameEn string `json:"nameEn"`
	// Unit represents the quotation measurement unit ("gram", "liter", "piece", "cylinder").
	Unit string `json:"unit"`
}

// BenchmarkIndex represents a market index (e.g. EGX30, TASI, DFMGI).
type BenchmarkIndex struct {
	// Symbol represents the market ticker symbol.
	Symbol string `json:"symbol"`
	// Code represents the normalized index code.
	Code string `json:"code"`
	// TvSymbol represents the TradingView symbol identifier.
	TvSymbol string `json:"tvSymbol"`
	// Exchange represents the host stock exchange ("EGX", "TADAWUL", "DFM", "ADX", "QSE", "KSE").
	Exchange string `json:"exchange"`
	// Country represents the 2-letter ISO country code.
	Country string `json:"country"`
	// Currency represents the trading currency of the exchange.
	Currency string `json:"currency"`
	// NameAr represents the Arabic index title.
	NameAr string `json:"nameAr"`
	// NameEn represents the English index title.
	NameEn string `json:"nameEn"`
	// Logo represents the icon image URL.
	Logo string `json:"logo"`
	// ConstituentSymbols lists individual equity tickers comprising the benchmark index.
	ConstituentSymbols []string `json:"constituentSymbols"`
}

// CatalogCounts provides high-level summary metrics of all platform inventory entities.
type CatalogCounts struct {
	// TotalBanks counts supported banking institutions.
	TotalBanks int `json:"totalBanks"`
	// TotalCurrencies counts supported ISO foreign currencies.
	TotalCurrencies int `json:"totalCurrencies"`
	// TotalCommodities counts precious metals and petroleum fuel varieties.
	TotalCommodities int `json:"totalCommodities"`
	// TotalIndices counts benchmark equity indices across domestic and GCC markets.
	TotalIndices int `json:"totalIndices"`
	// TotalSecurities counts individual equities cataloged across all regional exchanges.
	TotalSecurities int `json:"totalSecurities"`
}

// CatalogInventory represents the consolidated platform inventory returned by GET /v1/catalog.
type CatalogInventory struct {
	// Banks lists all registered banking institutions.
	Banks []Bank `json:"banks"`
	// Currencies lists supported exchange currencies.
	Currencies []Currency `json:"currencies"`
	// Commodities lists precious metals and fuel petroleum items.
	Commodities []CommodityDefinition `json:"commodities"`
	// Indices lists benchmark market indices.
	Indices []BenchmarkIndex `json:"indices"`
	// Counts contains aggregated entity count metrics.
	Counts CatalogCounts `json:"counts"`
}

// CatalogRegistry provides thread-safe indexed lookups for all platform catalog entities.
type CatalogRegistry struct {
	// inventory holds the root deserialized catalog inventory object.
	inventory CatalogInventory
	// bankByCode indexes banks by normalized uppercase code.
	bankByCode map[string]Bank
	// currencyByCode indexes currencies by normalized uppercase code.
	currencyByCode map[string]Currency
	// commodityByCode indexes commodities by normalized uppercase code.
	commodityByCode map[string]CommodityDefinition
	// indexByCode indexes benchmark indices by normalized uppercase code or symbol.
	indexByCode map[string]BenchmarkIndex
}

// globalCatalogInstance holds the singleton catalog registry.
var (
	// globalCatalogInstance stores the lazily loaded registry reference.
	globalCatalogInstance *CatalogRegistry
	// catalogOnce ensures atomic one-time initialization of the registry.
	catalogOnce sync.Once
)

// GetCatalogRegistry returns the initialized singleton CatalogRegistry instance.
func GetCatalogRegistry() *CatalogRegistry {
	// Execute initialization block once across all concurrent goroutines.
	catalogOnce.Do(func() {
		// Initialize empty registry struct with maps.
		reg := &CatalogRegistry{
			// Allocate bank hash map.
			bankByCode: make(map[string]Bank),
			// Allocate currency hash map.
			currencyByCode: make(map[string]Currency),
			// Allocate commodity hash map.
			commodityByCode: make(map[string]CommodityDefinition),
			// Allocate benchmark index hash map.
			indexByCode: make(map[string]BenchmarkIndex),
		}

		// Deserialize embedded catalog JSON data into inventory struct.
		if err := json.Unmarshal(rawCatalogData, &reg.inventory); err == nil {
			// Populate bank map by normalized uppercase code.
			for _, b := range reg.inventory.Banks {
				// Index bank by uppercase code.
				reg.bankByCode[strings.ToUpper(b.Code)] = b
			}
			// Populate currency map by normalized uppercase code.
			for _, c := range reg.inventory.Currencies {
				// Index currency by uppercase code.
				reg.currencyByCode[strings.ToUpper(c.Code)] = c
			}
			// Populate commodity map by normalized uppercase code.
			for _, com := range reg.inventory.Commodities {
				// Index commodity by uppercase code.
				reg.commodityByCode[strings.ToUpper(com.Code)] = com
			}
			// Populate benchmark index map by uppercase code and symbol.
			for _, idx := range reg.inventory.Indices {
				// Index by code.
				reg.indexByCode[strings.ToUpper(idx.Code)] = idx
				// Index by symbol.
				reg.indexByCode[strings.ToUpper(idx.Symbol)] = idx
			}
		}

		// Store the populated registry in global pointer.
		globalCatalogInstance = reg
	})

	// Return initialized singleton registry.
	return globalCatalogInstance
}

// GetInventory returns the complete catalog inventory for API response serialization.
func (r *CatalogRegistry) GetInventory() CatalogInventory {
	// Return the cached inventory value.
	return r.inventory
}

// FindBank retrieves bank metadata by uppercase code, returning true if found.
func (r *CatalogRegistry) FindBank(code string) (Bank, bool) {
	// Look up bank by normalized uppercase code.
	b, found := r.bankByCode[strings.ToUpper(strings.TrimSpace(code))]
	// Return bank entity and boolean found flag.
	return b, found
}

// FindCurrency retrieves currency metadata by uppercase code, returning true if found.
func (r *CatalogRegistry) FindCurrency(code string) (Currency, bool) {
	// Look up currency by normalized uppercase code.
	c, found := r.currencyByCode[strings.ToUpper(strings.TrimSpace(code))]
	// Return currency entity and boolean found flag.
	return c, found
}

// IsSupportedCurrency returns true if the provided currency code is known to the gateway.
func (r *CatalogRegistry) IsSupportedCurrency(code string) bool {
	// Check map existence for normalized currency code.
	_, found := r.FindCurrency(code)
	// Return existence result.
	return found
}
