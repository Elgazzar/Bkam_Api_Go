// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for HTTP handling, string normalization, and time manipulation.
import (
	// fmt formats error messages.
	"fmt"
	// net/http provides HTTP request and response writer interfaces.
	"net/http"
	// strings provides uppercase and trimming normalization.
	"strings"
	// time provides timestamping.
	"time"

	// chi provides URL parameter extraction.
	"github.com/go-chi/chi/v5"

	// domain defines core financial models and catalog registry.
	"bkam-api/internal/domain"
	// engine provides the LiveSnapshotCoordinator storing in-memory quotes.
	"bkam-api/internal/engine"
)

// RatesHandler handles foreign exchange rates and banking catalog HTTP endpoints.
type RatesHandler struct {
	// coordinator supplies in-memory live rates snapshots.
	coordinator *engine.LiveSnapshotCoordinator
	// historyEngine supplies historical date archives and multi-day range slices.
	historyEngine *engine.HistoryResolutionEngine
	// intradayStore supplies intraday tick series and session backtracking.
	intradayStore *engine.IntradayTimelineStore
	// catalog supplies bank and currency registry lookups.
	catalog *domain.CatalogRegistry
}

// NewRatesHandler constructs a new RatesHandler instance.
func NewRatesHandler(
	coordinator *engine.LiveSnapshotCoordinator,
	historyEngine *engine.HistoryResolutionEngine,
) *RatesHandler {
	// Initialize handler with dependencies.
	return &RatesHandler{
		// Assign coordinator.
		coordinator: coordinator,
		// Assign history engine.
		historyEngine: historyEngine,
		// Assign catalog registry.
		catalog: domain.GetCatalogRegistry(),
	}
}

// SetIntradayStore sets the optional IntradayTimelineStore dependency.
func (h *RatesHandler) SetIntradayStore(store *engine.IntradayTimelineStore) {
	// Assign store pointer.
	h.intradayStore = store
}

// GetCatalog handles GET /v1/catalog, returning static platform inventory metadata.
func (h *RatesHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	// Set 24-hour cache control header matching production specification.
	w.Header().Set("Cache-Control", "public, max-age=86400, stale-while-revalidate=3600")
	// Respond with complete catalog inventory JSON payload.
	RespondJSON(w, http.StatusOK, h.catalog.GetInventory())
}

// GetRates handles GET /v1/rates, supporting optional ?currency= and ?bank= query filtering.
func (h *RatesHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	// Extract optional currency query parameter.
	currParam := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("currency")))
	// Extract optional bank query parameter.
	bankParam := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("bank")))

	// If currency query parameter is provided, validate against catalog.
	if currParam != "" {
		// Check currency support.
		if !h.catalog.IsSupportedCurrency(currParam) {
			// Reject with 400 INVALID_FILTER error matching contract.
			RespondError(w, http.StatusBadRequest, "INVALID_FILTER", fmt.Sprintf("Currency '%s' is not supported by our platform.", currParam), nil)
			// Halt handler execution.
			return
		}
	}

	// If bank query parameter is provided, validate against catalog.
	if bankParam != "" {
		// Look up bank.
		if _, found := h.catalog.FindBank(bankParam); !found {
			// Reject with 400 INVALID_FILTER error.
			RespondError(w, http.StatusBadRequest, "INVALID_FILTER", fmt.Sprintf("Bank '%s' is not supported by our platform.", bankParam), nil)
			// Halt handler.
			return
		}
	}

	// Filter and serve rates snapshot.
	h.serveFilteredRates(w, currParam, bankParam)
}

// GetRatesByCurrency handles GET /v1/rates/{currency} path parameter endpoint.
func (h *RatesHandler) GetRatesByCurrency(w http.ResponseWriter, r *http.Request) {
	// Extract currency code from URL path parameter.
	currParam := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "currency")))

	// Validate path parameter length.
	if len(currParam) != 3 {
		// Reject malformed currency parameter.
		RespondError(w, http.StatusBadRequest, "INVALID_FILTER", "Invalid currency path parameter format. Must be a 3-character string.", nil)
		// Halt handler.
		return
	}

	// Verify currency code exists in official catalog.
	if !h.catalog.IsSupportedCurrency(currParam) {
		// Reject unsupported currency code with 400 INVALID_FILTER.
		RespondError(w, http.StatusBadRequest, "INVALID_FILTER", fmt.Sprintf("Currency '%s' is not supported by our platform.", currParam), nil)
		// Halt handler.
		return
	}

	// Filter and serve rates snapshot.
	h.serveFilteredRates(w, currParam, "")
}

// GetRatesByBank handles GET /v1/banks/{bankCode}/rates path parameter endpoint.
func (h *RatesHandler) GetRatesByBank(w http.ResponseWriter, r *http.Request) {
	// Extract bank code from URL path parameter.
	bankParam := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "bankCode")))
	// Extract optional currency query parameter.
	currParam := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("currency")))

	// Validate bank code in catalog.
	if _, found := h.catalog.FindBank(bankParam); !found {
		// Reject unknown bank code with 400 INVALID_FILTER.
		RespondError(w, http.StatusBadRequest, "INVALID_FILTER", fmt.Sprintf("Bank '%s' is not supported by our platform.", bankParam), nil)
		// Halt handler.
		return
	}

	// If currency query parameter provided, validate it too.
	if currParam != "" && !h.catalog.IsSupportedCurrency(currParam) {
		// Reject unsupported currency with 400 INVALID_FILTER.
		RespondError(w, http.StatusBadRequest, "INVALID_FILTER", fmt.Sprintf("Currency '%s' is not supported by our platform.", currParam), nil)
		// Halt handler.
		return
	}

	// Filter and serve rates snapshot.
	h.serveFilteredRates(w, currParam, bankParam)
}

// serveFilteredRates applies currency and bank filters to memory snapshot and writes HTTP response.
func (h *RatesHandler) serveFilteredRates(w http.ResponseWriter, currencyFilter, bankFilter string) {
	// Retrieve typed rates snapshot from in-memory coordinator.
	snapshot, exists := h.coordinator.GetRatesSnapshot()

	// If snapshot is not yet loaded in memory.
	if !exists || snapshot == nil {
		// Synthesize empty placeholder snapshot to preserve contract structure.
		emptySnap := domain.RateSnapshot{
			RequestedAt:    time.Now().UTC(),
			Freshness:      "stale",
			RequestFilters: make(map[string]string),
			Rates:          make([]domain.CurrencyRate, 0),
			Availability:   make([]domain.AvailabilityRecord, 0),
			Warnings:       []string{"Snapshot cache is warming."},
		}
		// Return 200 OK with empty snapshot.
		RespondJSON(w, http.StatusOK, emptySnap)
		// Finish.
		return
	}

	// Allocate filtered rates slice.
	filteredRates := make([]domain.CurrencyRate, 0, len(snapshot.Rates))
	// Allocate filters map.
	filters := make(map[string]string)

	// Set currency filter in map if present.
	if currencyFilter != "" {
		// Record filter.
		filters["currency"] = currencyFilter
	}
	// Set bank filter in map if present.
	if bankFilter != "" {
		// Record filter.
		filters["bank"] = bankFilter
	}

	// Iterate through all rates in memory snapshot.
	for _, rate := range snapshot.Rates {
		// If currency filter active, skip mismatching currencies.
		if currencyFilter != "" && !strings.EqualFold(rate.CurrencyCode, currencyFilter) {
			// Skip mismatch.
			continue
		}
		// If bank filter active, skip mismatching banks.
		if bankFilter != "" && !strings.EqualFold(rate.BankCode, bankFilter) {
			// Skip mismatch.
			continue
		}
		// Append matching rate item.
		filteredRates = append(filteredRates, rate)
	}

	// Construct filtered response envelope preserving root metadata.
	responseEnvelope := domain.RateSnapshot{
		// Processed timestamp.
		RequestedAt: time.Now().UTC(),
		// Freshness status.
		Freshness: snapshot.Freshness,
		// Applied query filters.
		RequestFilters: filters,
		// Filtered rates list.
		Rates: filteredRates,
		// Availability list.
		Availability: snapshot.Availability,
		// Operational warnings.
		Warnings: snapshot.Warnings,
	}

	// If filtered by single currency, include top-level currency field.
	if currencyFilter != "" {
		// Populate currency.
		responseEnvelope.Currency = currencyFilter
	}

	// Transmit JSON response.
	RespondJSON(w, http.StatusOK, responseEnvelope)
}

// GetRatesHistoryByDate handles GET /v1/rates/history/{date} single calendar date archive queries.
func (h *RatesHandler) GetRatesHistoryByDate(w http.ResponseWriter, r *http.Request) {
	// Extract date from URL path.
	dateParam := strings.TrimSpace(chi.URLParam(r, "date"))

	// Validate date format YYYY-MM-DD.
	if _, err := time.Parse("2006-01-02", dateParam); err != nil {
		// Reject with 400 INVALID_VALIDATION.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid date format. Expected YYYY-MM-DD.", nil)
		// Return.
		return
	}

	// If history engine not configured.
	if h.historyEngine == nil {
		// Return 404 NOT_FOUND.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Historical rates not found for %s", dateParam), nil)
		// Return.
		return
	}

	// Extract optional bank query parameter filter.
	bank := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("bank")))
	// If missing in query, check path parameter bankCode.
	if bank == "" {
		// Extract from path parameter.
		bank = strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "bankCode")))
	}
	// Extract optional currency query parameter filter.
	currency := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("currency")))

	// Resolve snapshot via history engine with query options.
	resolved, err := h.historyEngine.GetDateSnapshot("rates", dateParam, domain.RangeOptions{
		// Filter by bank code if provided.
		Bank: bank,
		// Filter by currency code if provided.
		Currency: currency,
	})
	// If error returned (e.g. future date).
	if err != nil {
		// Reject with 400 INVALID_VALIDATION.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", err.Error(), nil)
		// Return.
		return
	}

	// If resolved data is missing.
	if resolved == nil || resolved.Data == nil {
		// Return 404 NOT_FOUND.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Historical rates not found for %s", dateParam), nil)
		// Return.
		return
	}

	// Set dynamic Cache-Control header.
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", resolved.CacheTTLSeconds))
	// Transmit resolved payload.
	RespondJSON(w, http.StatusOK, resolved.Data)
}

// GetRatesHistoryRange handles GET /v1/rates/history/range multi-day date range queries.
func (h *RatesHandler) GetRatesHistoryRange(w http.ResponseWriter, r *http.Request) {
	// Extract start date query parameter.
	start := strings.TrimSpace(r.URL.Query().Get("start"))
	// Extract end date query parameter.
	end := strings.TrimSpace(r.URL.Query().Get("end"))
	// Extract optional bank filter.
	bank := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("bank")))
	// If missing in query, check path parameter bankCode.
	if bank == "" {
		// Extract from path parameter.
		bank = strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "bankCode")))
	}
	// Extract optional currency filter.
	currency := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("currency")))

	// Validate start date format.
	if _, err := time.Parse("2006-01-02", start); err != nil {
		// Reject with 400 INVALID_VALIDATION.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid range parameters. Dates must be formatted as YYYY-MM-DD, cannot be in the future, and start must be before or equal to end.", nil)
		// Return.
		return
	}

	// Validate end date format.
	if _, err := time.Parse("2006-01-02", end); err != nil {
		// Reject with 400 INVALID_VALIDATION.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid range parameters. Dates must be formatted as YYYY-MM-DD, cannot be in the future, and start must be before or equal to end.", nil)
		// Return.
		return
	}

	// If history engine not configured, return empty array.
	if h.historyEngine == nil {
		// Return empty list.
		RespondJSON(w, http.StatusOK, []any{})
		// Return.
		return
	}

	// Query multi-day historical range.
	results, err := h.historyEngine.GetDateRange("rates", domain.RangeOptions{
		StartDate: start,
		EndDate:   end,
		Bank:      bank,
		Currency:  currency,
	})
	// Check validation error.
	if err != nil {
		// Reject with 400 INVALID_VALIDATION.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", err.Error(), nil)
		// Return.
		return
	}

	// Cache header: 60s if end is today, 3600s if sealed past.
	w.Header().Set("Cache-Control", "public, max-age=3600")
	// Respond with results array.
	RespondJSON(w, http.StatusOK, results)
}

// GetRatesIntradayLatest handles GET /v1/rates/:currency/intraday/latest with optional ?bank= filter.
func (h *RatesHandler) GetRatesIntradayLatest(w http.ResponseWriter, r *http.Request) {
	// Extract currency parameter from URL path.
	currParam := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "currency")))
	// Extract optional bank query parameter.
	bankParam := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("bank")))

	// Validate currency code presence.
	if currParam == "" {
		// Reject missing currency with 400 INVALID_FILTER.
		RespondError(w, http.StatusBadRequest, "INVALID_FILTER", "Currency parameter is required.", nil)
		// Return early.
		return
	}

	// Validate currency against platform catalog.
	if !h.catalog.IsSupportedCurrency(currParam) {
		// Reject unsupported currency.
		RespondError(w, http.StatusBadRequest, "INVALID_FILTER", fmt.Sprintf("Currency '%s' is not supported by our platform.", currParam), nil)
		// Return early.
		return
	}

	// If bank filter provided, validate in catalog.
	if bankParam != "" {
		// Look up bank.
		if _, found := h.catalog.FindBank(bankParam); !found {
			// Reject unknown bank.
			RespondError(w, http.StatusBadRequest, "INVALID_FILTER", fmt.Sprintf("Bank '%s' is not supported by our platform.", bankParam), nil)
			// Return early.
			return
		}
	}

	// Default session date and freshness.
	sessionDate := time.Now().UTC().Format("2006-01-02")
	// Track whether session is today.
	isToday := true

	// If intraday store is configured, resolve authoritative session date.
	if h.intradayStore != nil {
		// Resolve session date.
		sessionDate, isToday = h.intradayStore.ResolveSessionDate("rates", "default")
	}

	// Extract intraday points.
	intradayPoints := make([]domain.IntradayPoint, 0)
	// If intraday store exists.
	if h.intradayStore != nil {
		// Retrieve points map for rates.
		if _, pointsMap, exists := h.intradayStore.GetIntraday("rates", "default"); exists && pointsMap != nil {
			// Unwrap if still nested under intradayPoints.
			if inner, ok := pointsMap["intradayPoints"].(map[string]any); ok && inner != nil {
				// Use inner points map.
				pointsMap = inner
			}
			// If bank filter supplied.
			if bankParam != "" {
				// Case-insensitive bank lookup.
				var bMap map[string]any
				// Search for matching bank key.
				for k, v := range pointsMap {
					// Check case-insensitive equality.
					if strings.EqualFold(k, bankParam) {
						// Extract bank points map.
						if m, ok := v.(map[string]any); ok {
							// Found bank map.
							bMap = m
							// Break loop.
							break
						}
					}
				}
				// If bank map found.
				if bMap != nil {
					// Lookup currency points in bank map.
					intradayPoints = h.intradayStore.LookupSymbolPoints(bMap, currParam)
				}
			} else {
				// Search benchmark banks first (nbe, cbe, cib, bm, banquemisr).
				benchmarkBanks := []string{"nbe", "cbe", "cib", "bm", "banquemisr"}
				// Loop over benchmark banks.
				for _, bCode := range benchmarkBanks {
					// Loop over points map.
					for k, v := range pointsMap {
						// Check matching benchmark bank.
						if strings.EqualFold(k, bCode) {
							// Cast to map.
							if m, ok := v.(map[string]any); ok {
								// Lookup currency points.
								pts := h.intradayStore.LookupSymbolPoints(m, currParam)
								// If points found.
								if len(pts) > 0 {
									// Store points.
									intradayPoints = pts
									// Break.
									break
								}
							}
						}
					}
					// If points found, stop benchmark search.
					if len(intradayPoints) > 0 {
						// Break.
						break
					}
				}
				// If still empty, search all bank entries.
				if len(intradayPoints) == 0 {
					// Iterate all bank points in dictionary.
					for _, v := range pointsMap {
						// Cast to map.
						if m, ok := v.(map[string]any); ok {
							// Lookup points.
							pts := h.intradayStore.LookupSymbolPoints(m, currParam)
							// If points found.
							if len(pts) > 0 {
								// Assign points.
								intradayPoints = pts
								// Break.
								break
							}
						}
					}
				}
			}
		}
	}

	// Initialize price metrics.
	var buyRate float64
	// Initialize sell rate.
	var sellRate float64
	// Initialize previous buy.
	var prevBuy float64
	// Initialize previous sell.
	var prevSell float64
	// Update timestamp string.
	updatedAt := time.Now().UTC().Format(time.RFC3339)

	// Fetch live snapshot for today.
	if snap, exists := h.coordinator.GetRatesSnapshot(); exists && snap != nil {
		// Update timestamp.
		updatedAt = snap.RequestedAt.Format(time.RFC3339)
		// Matching rates container.
		var matchedRates []domain.CurrencyRate
		// Filter rates.
		for _, rate := range snap.Rates {
			// Match currency.
			if strings.EqualFold(rate.CurrencyCode, currParam) {
				// If bank specified, match bank too.
				if bankParam != "" {
					// Compare bank.
					if strings.EqualFold(rate.BankCode, bankParam) {
						// Append matching.
						matchedRates = append(matchedRates, rate)
					}
				} else {
					// Append to consolidated.
					matchedRates = append(matchedRates, rate)
				}
			}
		}

		// If matching rates found.
		if len(matchedRates) > 0 {
			// Sum buy rates.
			var totalBuy float64
			// Sum sell rates.
			var totalSell float64
			// Previous buys.
			var prevBuys []float64
			// Previous sells.
			var prevSells []float64

			// Accumulate rates.
			for _, r := range matchedRates {
				// Accumulate buy.
				totalBuy += r.BuyRate
				// Accumulate sell.
				totalSell += r.SellRate

				// Check previous close buy.
				if r.YesterdayCloseBuy != nil && *r.YesterdayCloseBuy > 0 {
					// Append prev buy.
					prevBuys = append(prevBuys, *r.YesterdayCloseBuy)
				} else if r.OldBuyRate != nil && *r.OldBuyRate > 0 {
					// Append old buy.
					prevBuys = append(prevBuys, *r.OldBuyRate)
				}

				// Check previous close sell.
				if r.YesterdayCloseSell != nil && *r.YesterdayCloseSell > 0 {
					// Append prev sell.
					prevSells = append(prevSells, *r.YesterdayCloseSell)
				} else if r.OldSellRate != nil && *r.OldSellRate > 0 {
					// Append old sell.
					prevSells = append(prevSells, *r.OldSellRate)
				}
			}

			// Compute average buy rate.
			buyRate = roundTo(totalBuy/float64(len(matchedRates)), 2)
			// Compute average sell rate.
			sellRate = roundTo(totalSell/float64(len(matchedRates)), 2)

			// Compute average prev buy.
			if len(prevBuys) > 0 {
				// Sum.
				var sum float64
				// Loop.
				for _, v := range prevBuys {
					sum += v
				}
				// Avg.
				prevBuy = roundTo(sum/float64(len(prevBuys)), 2)
			} else {
				// Fallback to buy rate.
				prevBuy = buyRate
			}

			// Compute average prev sell.
			if len(prevSells) > 0 {
				// Sum.
				var sum float64
				// Loop.
				for _, v := range prevSells {
					sum += v
				}
				// Avg.
				prevSell = roundTo(sum/float64(len(prevSells)), 2)
			} else {
				// Fallback to sell rate.
				prevSell = sellRate
			}
		}
	}

	// Set dynamic cache control header: 60s for live today, 3600s for historical.
	if isToday {
		// Live 60s cache.
		w.Header().Set("Cache-Control", "public, max-age=60")
	} else {
		// Past session 3600s cache.
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	// Construct response payload.
	responsePayload := domain.RatesIntradayResponse{
		Currency:          currParam,
		Bank:              strings.ToLower(bankParam),
		Date:              sessionDate,
		Buy:               buyRate,
		Sell:              sellRate,
		PreviousBuyPrice:  prevBuy,
		PreviousSellPrice: prevSell,
		UpdatedAt:         updatedAt,
		Intraday:          intradayPoints,
	}

	// Transmit JSON response.
	RespondJSON(w, http.StatusOK, responsePayload)
}
