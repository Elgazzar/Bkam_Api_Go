// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for HTTP interfaces and time.
import (
	// fmt formats error messages.
	"fmt"
	// net/http provides response and request handling primitives.
	"net/http"
	// strings provides string normalization.
	"strings"
	// time provides timestamp generation.
	"time"

	// chi extracts URL path parameters.
	"github.com/go-chi/chi/v5"

	// domain defines precious metals and petroleum models.
	"bkam-api/internal/domain"
	// engine provides the LiveSnapshotCoordinator storing quotes in memory.
	"bkam-api/internal/engine"
)

// CommoditiesHandler manages endpoints for precious metals (gold) and petroleum products (gas).
type CommoditiesHandler struct {
	// coordinator supplies in-memory live snapshots.
	coordinator *engine.LiveSnapshotCoordinator
	// historyEngine supplies historical daily snapshots and ranges.
	historyEngine *engine.HistoryResolutionEngine
	// intradayStore supplies intraday tick sparkline series.
	intradayStore *engine.IntradayTimelineStore
}

// NewCommoditiesHandler constructs and returns a new CommoditiesHandler instance.
func NewCommoditiesHandler(
	coordinator *engine.LiveSnapshotCoordinator,
	historyEngine *engine.HistoryResolutionEngine,
) *CommoditiesHandler {
	// Return handler struct.
	return &CommoditiesHandler{
		// Assign coordinator.
		coordinator: coordinator,
		// Assign history engine.
		historyEngine: historyEngine,
	}
}

// SetIntradayStore injects the optional IntradayTimelineStore dependency.
func (h *CommoditiesHandler) SetIntradayStore(store *engine.IntradayTimelineStore) {
	// Assign intraday store pointer.
	h.intradayStore = store
}

// GetGold handles GET /v1/gold, returning the latest gold prices across all karats.
func (h *CommoditiesHandler) GetGold(w http.ResponseWriter, r *http.Request) {
	// Retrieve typed gold snapshot from in-memory coordinator.
	snapshot, exists := h.coordinator.GetGoldSnapshot()

	// If snapshot is missing from memory cache.
	if !exists || snapshot == nil {
		// Synthesize fallback empty gold snapshot.
		fallback := domain.GoldSnapshot{
			// Authoritative fallback source URL.
			Source: "https://market.isagha.com/prices",
			// Pricing currency.
			Currency: "EGP",
			// Official logo asset URL.
			Logo: "https://bekamfeelbank.web.app/images/gold.png",
			// Freshness indicator.
			Freshness: "stale",
			// Distributed trace identifier.
			SourceTraceID: "cold-boot",
			// Update timestamp.
			UpdatedAt: time.Now().UTC(),
			// Ingestion timestamp.
			ScrapedAt: time.Now().UTC(),
			// Empty rates map.
			Rates: make(map[string]domain.GoldKaratPrice),
		}
		// Return 200 OK with fallback.
		RespondJSON(w, http.StatusOK, fallback)
		// Return.
		return
	}

	// Create shallow copy to normalize envelope fields without mutating memory state.
	res := *snapshot
	// Verify that source string begins with http:// or https://.
	if !strings.HasPrefix(res.Source, "http://") && !strings.HasPrefix(res.Source, "https://") {
		// Set standard public source URL.
		res.Source = "https://market.isagha.com/prices"
	}
	// Verify currency string length is 3 characters.
	if len(res.Currency) != 3 {
		// Default to EGP.
		res.Currency = "EGP"
	}
	// Verify logo is populated with a valid URL.
	if res.Logo == "" {
		// Assign asset logo URL.
		res.Logo = "https://bekamfeelbank.web.app/images/gold.png"
	}
	// Verify updatedAt is non-zero.
	if res.UpdatedAt.IsZero() {
		// Assign current UTC time.
		res.UpdatedAt = time.Now().UTC()
	}
	// Verify scrapedAt is non-zero.
	if res.ScrapedAt.IsZero() {
		// Assign current UTC time.
		res.ScrapedAt = time.Now().UTC()
	}

	// Transmit normalized snapshot JSON.
	RespondJSON(w, http.StatusOK, res)
}

// GetGas handles GET /v1/gas, returning latest petroleum fuel prices.
func (h *CommoditiesHandler) GetGas(w http.ResponseWriter, r *http.Request) {
	// Retrieve typed gas snapshot from in-memory coordinator.
	snapshot, exists := h.coordinator.GetGasSnapshot()

	// If snapshot is missing from memory cache.
	if !exists || snapshot == nil {
		// Synthesize fallback gas snapshot.
		fallback := domain.GasSnapshot{
			// Petroleum authority source URL.
			Source: "https://www.petroleum.gov.eg",
			// Pricing currency.
			Currency: "EGP",
			// Official petroleum logo asset URL.
			Logo: "https://bekamfeelbank.web.app/images/gas.png",
			// Freshness indicator.
			Freshness: "stale",
			// Distributed trace identifier.
			SourceTraceID: "cold-boot",
			// Update timestamp.
			UpdatedAt: time.Now().UTC(),
			// Ingestion timestamp.
			ScrapedAt: time.Now().UTC(),
			// Empty rates map.
			Rates: make(map[string]domain.GasPriceItem),
		}
		// Return 200 OK with fallback.
		RespondJSON(w, http.StatusOK, fallback)
		// Return.
		return
	}

	// Create shallow copy to normalize envelope fields.
	res := *snapshot
	// Verify that source string begins with http:// or https://.
	if !strings.HasPrefix(res.Source, "http://") && !strings.HasPrefix(res.Source, "https://") {
		// Set standard petroleum portal source URL.
		res.Source = "https://www.petroleum.gov.eg"
	}
	// Verify currency string length is 3 characters.
	if len(res.Currency) != 3 {
		// Default to EGP.
		res.Currency = "EGP"
	}
	// Verify logo is populated with a valid URL.
	if res.Logo == "" {
		// Assign petroleum asset logo URL.
		res.Logo = "https://bekamfeelbank.web.app/images/gas.png"
	}
	// Verify updatedAt is non-zero.
	if res.UpdatedAt.IsZero() {
		// Assign current UTC time.
		res.UpdatedAt = time.Now().UTC()
	}
	// Verify scrapedAt is non-zero.
	if res.ScrapedAt.IsZero() {
		// Assign current UTC time.
		res.ScrapedAt = time.Now().UTC()
	}
	// Allocate container map for normalized fuel rates.
	normalizedRates := make(map[string]domain.GasPriceItem, len(res.Rates))
	// Loop over each petroleum product.
	for k, item := range res.Rates {
		// If unit is blank, assign standard retail volume unit.
		if item.Unit == "" {
			// Default unit to liter.
			item.Unit = "liter"
		}
		// Store normalized item.
		normalizedRates[k] = item
	}
	// Reassign normalized rates map.
	res.Rates = normalizedRates

	// Transmit normalized snapshot JSON.
	RespondJSON(w, http.StatusOK, res)
}

// GetGoldHistoryByDate handles GET /v1/gold/history/{date} queries for a single calendar date.
func (h *CommoditiesHandler) GetGoldHistoryByDate(w http.ResponseWriter, r *http.Request) {
	// Extract date string from URL route path parameter.
	dateParam := strings.TrimSpace(chi.URLParam(r, "date"))

	// Validate date format YYYY-MM-DD.
	if _, err := time.Parse("2006-01-02", dateParam); err != nil {
		// Reject malformed dates with HTTP 400 INVALID_VALIDATION.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid date format. Expected YYYY-MM-DD.", nil)
		// Return early.
		return
	}

	// If history engine reference is uninitialized.
	if h.historyEngine == nil {
		// Return HTTP 404 NOT_FOUND.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Historical gold rates not found for %s", dateParam), nil)
		// Return early.
		return
	}

	// Extract optional karat query parameter filter.
	karat := strings.TrimSpace(r.URL.Query().Get("karat"))

	// Query historical date snapshot from history engine.
	resolved, err := h.historyEngine.GetDateSnapshot("gold", dateParam, domain.RangeOptions{
		// Filter by karat denomination.
		Karat: karat,
	})
	// Check for validation errors such as future dates.
	if err != nil {
		// Reject invalid queries with HTTP 400 INVALID_VALIDATION.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", err.Error(), nil)
		// Return early.
		return
	}

	// If no data exists for the requested date.
	if resolved == nil || resolved.Data == nil {
		// Return HTTP 404 NOT_FOUND.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Historical gold rates not found for %s", dateParam), nil)
		// Return early.
		return
	}

	// Set dynamic Cache-Control header based on market schedule.
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", resolved.CacheTTLSeconds))
	// Transmit JSON response.
	RespondJSON(w, http.StatusOK, resolved.Data)
}

// GetGoldHistoryRange handles GET /v1/gold/history/range queries across a multi-day range.
func (h *CommoditiesHandler) GetGoldHistoryRange(w http.ResponseWriter, r *http.Request) {
	// Extract start date from query parameters.
	start := strings.TrimSpace(r.URL.Query().Get("start"))
	// Extract end date from query parameters.
	end := strings.TrimSpace(r.URL.Query().Get("end"))

	// Validate start date format.
	if _, err := time.Parse("2006-01-02", start); err != nil {
		// Reject invalid start date with HTTP 400.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid range parameters. Dates must be formatted as YYYY-MM-DD, cannot be in the future, and start must be before or equal to end.", nil)
		// Return early.
		return
	}

	// Validate end date format.
	if _, err := time.Parse("2006-01-02", end); err != nil {
		// Reject invalid end date with HTTP 400.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid range parameters. Dates must be formatted as YYYY-MM-DD, cannot be in the future, and start must be before or equal to end.", nil)
		// Return early.
		return
	}

	// If history engine is uninitialized.
	if h.historyEngine == nil {
		// Return empty JSON list.
		RespondJSON(w, http.StatusOK, []any{})
		// Return early.
		return
	}

	// Extract optional karat filter from query parameters.
	karat := strings.TrimSpace(r.URL.Query().Get("karat"))

	// Retrieve historical data range from history engine.
	results, err := h.historyEngine.GetDateRange("gold", domain.RangeOptions{
		// Starting date.
		StartDate: start,
		// Concluding date.
		EndDate: end,
		// Optional karat filter.
		Karat: karat,
	})
	// Check for range errors.
	if err != nil {
		// Reject invalid range with HTTP 400.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", err.Error(), nil)
		// Return early.
		return
	}

	// Set standard Cache-Control header.
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Assemble standardized range response payload matching public gateway contract.
	response := map[string]any{
		// Starting date string.
		"startDate": start,
		// Concluding date string.
		"endDate": end,
		// Historical range records array.
		"history": results,
	}
	// Transmit results response envelope.
	RespondJSON(w, http.StatusOK, response)
}

// GetGoldIntradayLatest handles GET /v1/gold/intraday/latest returning unified gold karats, pound, and international intraday series.
func (h *CommoditiesHandler) GetGoldIntradayLatest(w http.ResponseWriter, r *http.Request) {
	// Default session date and freshness.
	sessionDate := time.Now().UTC().Format("2006-01-02")
	// Track whether session is today.
	isToday := true

	// If intraday store is available.
	if h.intradayStore != nil {
		// Resolve session date.
		sessionDate, isToday = h.intradayStore.ResolveSessionDate("gold", "default")
	}

	// Points dictionary container.
	var pointsMap map[string]any
	// If intraday store available.
	if h.intradayStore != nil {
		// Retrieve points map for gold.
		_, pointsMap, _ = h.intradayStore.GetIntraday("gold", "default")
	}

	// Lookup points for 24k.
	p24 := h.lookupPoints(pointsMap, "24")
	// Lookup points for 21k.
	p21 := h.lookupPoints(pointsMap, "21")
	// Lookup points for 18k.
	p18 := h.lookupPoints(pointsMap, "18")
	// Lookup points for pound.
	pPound := h.lookupPoints(pointsMap, "pound")
	// Lookup points for international.
	pIntl := h.lookupPoints(pointsMap, "international")

	// Retrieve latest gold snapshot from coordinator.
	snap, _ := h.coordinator.GetGoldSnapshot()
	// Update timestamp string.
	updatedAt := time.Now().UTC().Format(time.RFC3339)
	// If snapshot exists.
	if snap != nil {
		// Format update timestamp.
		updatedAt = snap.UpdatedAt.Format(time.RFC3339)
	}

	// Helper to extract karat quotes.
	getQuote := func(k string) domain.GoldKaratIntradayItem {
		// Default item.
		item := domain.GoldKaratIntradayItem{Intraday: []domain.IntradayPoint{}}
		// If snapshot exists.
		if snap != nil && snap.Rates != nil {
			// Look up rate key e.g. "21k" or "21".
			rKey := k
			// If digit, append k.
			if k == "24" || k == "21" || k == "18" {
				// Format with k.
				rKey = k + "k"
			}
			// Look up rate item.
			if p, ok := snap.Rates[rKey]; ok {
				// Set buy.
				item.Buy = p.Buy
				// Set sell.
				item.Sell = p.Sell
				// Set prev buy.
				if p.YesterdayBuyPrice > 0 {
					item.PreviousBuyPrice = p.YesterdayBuyPrice
				} else {
					item.PreviousBuyPrice = p.Buy
				}
				// Set prev sell.
				if p.YesterdaySellPrice > 0 {
					item.PreviousSellPrice = p.YesterdaySellPrice
				} else if p.OldSell != nil && *p.OldSell > 0 {
					item.PreviousSellPrice = *p.OldSell
				} else {
					item.PreviousSellPrice = p.Sell
				}
			}
		}
		// Return populated item.
		return item
	}

	// Extract 24k quote.
	q24 := getQuote("24")
	// Set intraday points.
	q24.Intraday = p24

	// Extract 21k quote.
	q21 := getQuote("21")
	// Set intraday points.
	q21.Intraday = p21

	// Extract 18k quote.
	q18 := getQuote("18")
	// Set intraday points.
	q18.Intraday = p18

	// Extract pound quote.
	qPound := getQuote("pound")
	// Set intraday points.
	qPound.Intraday = pPound

	// Extract international quote.
	var qIntl domain.GoldInternationalIntradayItem
	// Set currency.
	qIntl.Currency = "USD"
	// Set points.
	qIntl.Intraday = pIntl
	// If international quote exists in snapshot.
	if snap != nil && snap.International != nil {
		// Set buy.
		qIntl.Buy = snap.International.Price
		// Set sell.
		qIntl.Sell = snap.International.Price
		// Set prev buy.
		qIntl.PreviousBuyPrice = snap.International.Price
		// Set prev sell.
		qIntl.PreviousSellPrice = snap.International.Price
		// Set currency if present.
		if snap.International.Currency != "" {
			qIntl.Currency = snap.International.Currency
		}
	}

	// Construct response payload.
	responsePayload := domain.GoldIntradayResponse{
		Date:      sessionDate,
		Currency:  "EGP",
		UpdatedAt: updatedAt,
		Karats: map[string]domain.GoldKaratIntradayItem{
			"24": q24,
			"21": q21,
			"18": q18,
		},
		Pound:         qPound,
		International: qIntl,
	}

	// Set dynamic Cache-Control header: 60s for live today, 3600s for past session.
	if isToday {
		// 60s cache.
		w.Header().Set("Cache-Control", "public, max-age=60")
	} else {
		// 3600s cache.
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	// Transmit JSON response.
	RespondJSON(w, http.StatusOK, responsePayload)
}

// GetGoldKaratIntradayLatest handles GET /v1/gold/:karat/intraday/latest returning single karat intraday series.
func (h *CommoditiesHandler) GetGoldKaratIntradayLatest(w http.ResponseWriter, r *http.Request) {
	// Extract karat parameter from URL path.
	rawKarat := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "karat")))

	// Normalize target item identifier.
	targetItem := ""
	// Match karat.
	switch rawKarat {
	case "24", "24k":
		targetItem = "24"
	case "21", "21k":
		targetItem = "21"
	case "18", "18k":
		targetItem = "18"
	case "pound", "gold_pound":
		targetItem = "pound"
	case "international", "intl", "ounce", "oz":
		targetItem = "international"
	default:
		// Reject unrecognized gold item with HTTP 400.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", fmt.Sprintf("Invalid gold item '%s'. Must be one of: 24, 21, 18, pound, international.", rawKarat), nil)
		// Return early.
		return
	}

	// Default session date and freshness.
	sessionDate := time.Now().UTC().Format("2006-01-02")
	// Track whether session is today.
	isToday := true

	// If intraday store available.
	if h.intradayStore != nil {
		// Resolve session date.
		sessionDate, isToday = h.intradayStore.ResolveSessionDate("gold", "default")
	}

	// Points container.
	var pointsMap map[string]any
	// If intraday store available.
	if h.intradayStore != nil {
		// Retrieve points map for gold.
		_, pointsMap, _ = h.intradayStore.GetIntraday("gold", "default")
	}

	// Lookup points for target item.
	points := h.lookupPoints(pointsMap, targetItem)

	// Retrieve latest gold snapshot from coordinator.
	snap, _ := h.coordinator.GetGoldSnapshot()
	// Update timestamp string.
	updatedAt := time.Now().UTC().Format(time.RFC3339)
	// Currency denomination.
	currency := "EGP"
	// Buy rate.
	var buyRate float64
	// Sell rate.
	var sellRate float64
	// Previous buy.
	var prevBuy float64
	// Previous sell.
	var prevSell float64

	// If snapshot exists.
	if snap != nil {
		// Set update timestamp.
		updatedAt = snap.UpdatedAt.Format(time.RFC3339)
		// Handle international item.
		if targetItem == "international" {
			// International currency.
			currency = "USD"
			// If international quote recorded.
			if snap.International != nil {
				buyRate = snap.International.Price
				sellRate = snap.International.Price
				prevBuy = snap.International.Price
				prevSell = snap.International.Price
				if snap.International.Currency != "" {
					currency = snap.International.Currency
				}
			}
		} else {
			// Karat lookup key.
			rKey := targetItem
			if targetItem == "24" || targetItem == "21" || targetItem == "18" {
				rKey = targetItem + "k"
			}
			// Look up rate.
			if p, ok := snap.Rates[rKey]; ok {
				buyRate = p.Buy
				sellRate = p.Sell
				if p.YesterdayBuyPrice > 0 {
					prevBuy = p.YesterdayBuyPrice
				} else {
					prevBuy = p.Buy
				}
				if p.YesterdaySellPrice > 0 {
					prevSell = p.YesterdaySellPrice
				} else if p.OldSell != nil && *p.OldSell > 0 {
					prevSell = *p.OldSell
				} else {
					prevSell = p.Sell
				}
			}
		}
	}

	// Construct single item response payload.
	responsePayload := domain.GoldSingleIntradayResponse{
		Item:              targetItem,
		Date:              sessionDate,
		Currency:          currency,
		Buy:               buyRate,
		Sell:              sellRate,
		PreviousBuyPrice:  prevBuy,
		PreviousSellPrice: prevSell,
		UpdatedAt:         updatedAt,
		Intraday:          points,
	}

	// Set dynamic Cache-Control header.
	if isToday {
		w.Header().Set("Cache-Control", "public, max-age=60")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	// Transmit JSON response.
	RespondJSON(w, http.StatusOK, responsePayload)
}

// lookupPoints safely searches an intraday points map using the store helper.
func (h *CommoditiesHandler) lookupPoints(pointsMap map[string]any, symbol string) []domain.IntradayPoint {
	// If intraday store is nil or map is nil.
	if h.intradayStore == nil || pointsMap == nil {
		// Return empty slice.
		return []domain.IntradayPoint{}
	}
	// Delegate lookup to store helper.
	return h.intradayStore.LookupSymbolPoints(pointsMap, symbol)
}
