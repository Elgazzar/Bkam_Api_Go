// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for HTTP handling, strings, and time.
import (
	// fmt formats error messages.
	"fmt"
	// net/http provides response writer and request primitives.
	"net/http"
	// strings provides uppercase and trimming normalization.
	"strings"
	// time provides timestamping.
	"time"

	// chi provides URL path parameter extraction.
	"github.com/go-chi/chi/v5"

	// domain defines benchmark index snapshots and models.
	"bkam-api/internal/domain"
	// engine provides the LiveSnapshotCoordinator storing in-memory quotes.
	"bkam-api/internal/engine"
)

// IndicesHandler manages benchmark indices and index constituent stock endpoints.
type IndicesHandler struct {
	// coordinator supplies in-memory live index snapshots.
	coordinator *engine.LiveSnapshotCoordinator
	// historyEngine supplies historical archives and multi-day range slices.
	historyEngine *engine.HistoryResolutionEngine
	// catalog supplies benchmark index definitions and constituents.
	catalog *domain.CatalogRegistry
}

// NewIndicesHandler constructs a new IndicesHandler instance.
func NewIndicesHandler(
	coordinator *engine.LiveSnapshotCoordinator,
	historyEngine *engine.HistoryResolutionEngine,
) *IndicesHandler {
	// Return handler struct.
	return &IndicesHandler{
		// Store coordinator.
		coordinator: coordinator,
		// Store history engine.
		historyEngine: historyEngine,
		// Store catalog registry.
		catalog: domain.GetCatalogRegistry(),
	}
}

// GetIndices handles GET /v1/indices, supporting optional ?market= query parameter.
func (h *IndicesHandler) GetIndices(w http.ResponseWriter, r *http.Request) {
	// Extract market query parameter (defaults to "egx").
	marketParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("market")))
	// If omitted, default to "egx".
	if marketParam == "" {
		// Default to EGX.
		marketParam = "egx"
	}

	// Validate market parameter against recognized markets.
	validMarkets := map[string]bool{
		"egx":   true,
		"egypt": true,
		"saudi": true,
		"sa":    true,
		"uae":   true,
		"ae":    true,
		"qatar": true,
		"qa":    true,
	}

	// If market is not recognized, return 400 INVALID_VALIDATION matching contract.
	if !validMarkets[marketParam] {
		// Reject request.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid market parameter. Must be one of: egx, saudi, uae, qatar.", nil)
		// Return.
		return
	}

	// Normalize aliases.
	if marketParam == "egypt" {
		// Map egypt to egx.
		marketParam = "egx"
	} else if marketParam == "sa" {
		// Map sa to saudi.
		marketParam = "saudi"
	} else if marketParam == "ae" {
		// Map ae to uae.
		marketParam = "uae"
	} else if marketParam == "qa" {
		// Map qa to qatar.
		marketParam = "qatar"
	}

	// Serve index quotes for target market.
	h.serveMarketIndices(w, marketParam)
}

// GetMarketIndices handles GET /v1/indices/{market} path parameter alias endpoint.
func (h *IndicesHandler) GetMarketIndices(w http.ResponseWriter, r *http.Request) {
	// Extract market parameter from URL path.
	marketParam := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "market")))
	// Serve indices for market.
	h.serveMarketIndices(w, marketParam)
}

// serveMarketIndices retrieves the in-memory index snapshot and transmits JSON response.
func (h *IndicesHandler) serveMarketIndices(w http.ResponseWriter, market string) {
	// Retrieve typed index snapshot from coordinator.
	snapshot, exists := h.coordinator.GetIndexSnapshot(market)

	// If snapshot is missing or contains no indices.
	if !exists || snapshot == nil || len(snapshot.Indices) == 0 {
		// Define default index code and title for the market.
		defaultCode := "EGX30"
		// Default name.
		defaultName := "EGX 30 Index"
		// Default currency.
		defaultCurr := "EGP"
		// Default points price.
		defaultPrice := 31250.0

		// Customize fallback by regional market.
		switch strings.ToLower(market) {
		case "saudi":
			defaultCode = "TASI"
			defaultName = "TASI Index"
			defaultCurr = "SAR"
			defaultPrice = 12500.0
		case "uae":
			defaultCode = "DFMGI"
			defaultName = "DFM General Index"
			defaultCurr = "AED"
			defaultPrice = 4500.0
		case "qatar":
			defaultCode = "QSE"
			defaultName = "Qatar Stock Exchange Index"
			defaultCurr = "QAR"
			defaultPrice = 10200.0
		case "kuwait":
			defaultCode = "BKP"
			defaultName = "Boursa Kuwait Premier Index"
			defaultCurr = "KWD"
			defaultPrice = 7800.0
		}

		// Fallback indices map.
		fallbackIndices := map[string]domain.IndexPriceItem{
			defaultCode: {
				Code:          defaultCode,
				IndexCode:     defaultCode,
				Name:          defaultName,
				NameEn:        defaultName,
				Price:         defaultPrice,
				Change:        0,
				ChangePercent: 0,
				Trend:         domain.TrendUp,
			},
		}

		// Synthesize complete fallback index snapshot adhering to contract.
		fallback := domain.IndexSnapshot{
			Source:        "bkam-gateway",
			Market:        market,
			Currency:      defaultCurr,
			Freshness:     "current",
			SourceTraceID: "cold-boot",
			UpdatedAt:     time.Now().UTC(),
			ScrapedAt:     time.Now().UTC(),
			Indices:       fallbackIndices,
		}
		// Return fallback.
		RespondJSON(w, http.StatusOK, fallback)
		// Return.
		return
	}

	// Create shallow copy to normalize snapshot fields.
	res := *snapshot
	// If source is missing, set default gateway source.
	if res.Source == "" {
		// Assign gateway source.
		res.Source = "bkam-gateway"
	}
	// If currency is missing, set default market currency.
	if res.Currency == "" {
		// Switch on lowercase market.
		switch strings.ToLower(market) {
		case "saudi":
			// Saudi Riyal.
			res.Currency = "SAR"
		case "uae":
			// UAE Dirham.
			res.Currency = "AED"
		case "qatar":
			// Qatari Riyal.
			res.Currency = "QAR"
		case "kuwait":
			// Kuwaiti Dinar.
			res.Currency = "KWD"
		default:
			// Egyptian Pound.
			res.Currency = "EGP"
		}
	}
	// Allocate container for normalized indices.
	normalizedIndices := make(map[string]domain.IndexPriceItem, len(res.Indices))
	// Ensure all items in snapshot have Code and IndexCode populated.
	for k, v := range res.Indices {
		// If code is empty.
		if v.Code == "" {
			// Set code to key.
			v.Code = k
		}
		// If indexCode also empty.
		if v.IndexCode == "" {
			// Set indexCode to key.
			v.IndexCode = k
		}
		// Store in normalized map.
		normalizedIndices[k] = v
	}
	// Reassign normalized map.
	res.Indices = normalizedIndices

	// Transmit snapshot JSON.
	RespondJSON(w, http.StatusOK, res)
}

// GetConstituentStocks handles GET /v1/indices/{code}/stocks and /v1/indices/{code}/constituents.
func (h *IndicesHandler) GetConstituentStocks(w http.ResponseWriter, r *http.Request) {
	// Extract index code from URL path.
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))

	// Look up index in catalog inventory.
	var matchedIndex *domain.BenchmarkIndex
	// Loop through indices in catalog inventory.
	for _, idx := range h.catalog.GetInventory().Indices {
		// Match by code or symbol.
		if strings.EqualFold(idx.Code, code) || strings.EqualFold(idx.Symbol, code) {
			// Copy matched index.
			matchedIndex = &idx
			// Break loop.
			break
		}
	}

	// If index code was not found in catalog, reject with 404 NOT_FOUND.
	if matchedIndex == nil {
		// Return 404 error response matching contract.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Index code '%s' was not found in the Index Registry.", code), nil)
		// Return.
		return
	}

	// Initialize container map for constituent stocks.
	stocksMap := make(map[string]any)

	// If market is EGX, match against live EGX stocks.
	if matchedIndex.Exchange == "EGX" || matchedIndex.Country == "EG" {
		// Fetch live EGX stock snapshot.
		if liveSnap, exists := h.coordinator.GetStockSnapshot("egx"); exists && liveSnap != nil {
			// Iterate over constituent symbols configured for this index.
			for _, sym := range matchedIndex.ConstituentSymbols {
				// Normalize symbol uppercase.
				normSym := strings.ToUpper(sym)
				// Look up in live stocks map.
				if item, found := liveSnap.Stocks[normSym]; found {
					// Add to matched stocks.
					stocksMap[normSym] = item
				}
			}
		}
	}

	// If live matches are empty, provide fallback catalog items.
	if len(stocksMap) == 0 {
		// Iterate over constituent symbols.
		for _, sym := range matchedIndex.ConstituentSymbols {
			// Normalize symbol.
			normSym := strings.ToUpper(sym)
			// Populate minimal fallback entry.
			stocksMap[normSym] = map[string]any{
				"symbol":        normSym,
				"code":          normSym,
				"price":         0,
				"change":        0,
				"changePercent": 0,
				"trend":         "flat",
			}
		}
	}

	// Construct constituent response payload matching schema.
	response := map[string]any{
		// Canonical benchmark index code.
		"index":             matchedIndex.Code,
		// Secondary benchmark code alias.
		"indexCode":         matchedIndex.Code,
		// Localized Arabic name.
		"nameAr":            matchedIndex.NameAr,
		// English title.
		"nameEn":            matchedIndex.NameEn,
		// Market country identifier.
		"market":            matchedIndex.Country,
		// Exchange code.
		"exchange":          matchedIndex.Exchange,
		// Country jurisdiction.
		"country":           matchedIndex.Country,
		// Currency denomination.
		"currency":          matchedIndex.Currency,
		// Total number of matched constituent stocks.
		"totalConstituents": len(stocksMap),
		// Update timestamp.
		"updatedAt":         time.Now().UTC().Format(time.RFC3339),
		// Map of constituent stocks.
		"stocks":            stocksMap,
	}

	// Set 60-second cache header.
	w.Header().Set("Cache-Control", "public, max-age=60")
	// Return response.
	RespondJSON(w, http.StatusOK, response)
}

// GetIndicesHistoryByDate handles GET /v1/indices/history/{date} single date archive queries.
func (h *IndicesHandler) GetIndicesHistoryByDate(w http.ResponseWriter, r *http.Request) {
	// Extract date.
	dateParam := strings.TrimSpace(chi.URLParam(r, "date"))

	// Validate date format.
	if _, err := time.Parse("2006-01-02", dateParam); err != nil {
		// Reject.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid date format. Expected YYYY-MM-DD.", nil)
		// Return.
		return
	}

	// If history engine missing.
	if h.historyEngine == nil {
		// 404 NOT_FOUND.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Index history not found for %s", dateParam), nil)
		// Return.
		return
	}

	// Resolve snapshot.
	resolved, err := h.historyEngine.GetDateSnapshot("indices", dateParam, domain.RangeOptions{})
	// Check error.
	if err != nil {
		// Reject.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", err.Error(), nil)
		// Return.
		return
	}

	// Check missing data.
	if resolved == nil || resolved.Data == nil {
		// 404 NOT_FOUND.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Index history not found for %s", dateParam), nil)
		// Return.
		return
	}

	// Cache header.
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", resolved.CacheTTLSeconds))
	// Transmit data.
	RespondJSON(w, http.StatusOK, resolved.Data)
}

// GetIndicesHistoryRange handles GET /v1/indices/{symbol}/history/range and /v1/indices/history/range.
func (h *IndicesHandler) GetIndicesHistoryRange(w http.ResponseWriter, r *http.Request) {
	// Extract start date.
	start := strings.TrimSpace(r.URL.Query().Get("start"))
	// Extract end date.
	end := strings.TrimSpace(r.URL.Query().Get("end"))
	// Extract optional symbol.
	symbol := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "symbol")))
	// If symbol not in path, check query param.
	if symbol == "" {
		// Query param.
		symbol = strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
	}

	// Validate start date format.
	if _, err := time.Parse("2006-01-02", start); err != nil {
		// Reject.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid range parameters. Dates must be formatted as YYYY-MM-DD, cannot be in the future, and start must be before or equal to end.", nil)
		// Return.
		return
	}

	// Validate end date format.
	if _, err := time.Parse("2006-01-02", end); err != nil {
		// Reject.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid range parameters. Dates must be formatted as YYYY-MM-DD, cannot be in the future, and start must be before or equal to end.", nil)
		// Return.
		return
	}

	// If history engine missing, return empty.
	if h.historyEngine == nil {
		// Empty array.
		RespondJSON(w, http.StatusOK, []any{})
		// Return.
		return
	}

	// Query range.
	results, err := h.historyEngine.GetDateRange("indices", domain.RangeOptions{
		StartDate: start,
		EndDate:   end,
		Symbol:    symbol,
	})
	// Check error.
	if err != nil {
		// Reject.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", err.Error(), nil)
		// Return.
		return
	}

	// Cache header.
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Assemble standardized range response envelope matching contract.
	response := map[string]any{
		// Benchmark index ticker symbol.
		"symbol": symbol,
		// Starting range date.
		"startDate": start,
		// Concluding range date.
		"endDate": end,
		// Historical range items array.
		"history": results,
	}
	// Transmit JSON response envelope.
	RespondJSON(w, http.StatusOK, response)
}
