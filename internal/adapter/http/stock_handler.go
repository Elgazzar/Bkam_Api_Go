// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for HTTP handling and timestamps.
import (
	// fmt formats error messages.
	"fmt"
	// net/http provides response writer and request primitives.
	"net/http"
	// strings provides string normalization.
	"strings"
	// time provides timestamp generation.
	"time"

	// chi extracts URL path parameters.
	"github.com/go-chi/chi/v5"

	// domain defines equity snapshots and regional catalog.
	"bkam-api/internal/domain"
	// engine provides the LiveSnapshotCoordinator storing quotes in memory.
	"bkam-api/internal/engine"
)

// StockHandler manages stock market endpoints across Egyptian (EGX) and regional GCC markets.
type StockHandler struct {
	// coordinator supplies in-memory live snapshots.
	coordinator *engine.LiveSnapshotCoordinator
	// historyEngine supplies historical archives and multi-day range slices.
	historyEngine *engine.HistoryResolutionEngine
	// intradayStore supplies intraday tick series and session backtracking.
	intradayStore *engine.IntradayTimelineStore
}

// NewStockHandler constructs a new StockHandler instance.
func NewStockHandler(
	coordinator *engine.LiveSnapshotCoordinator,
	historyEngine *engine.HistoryResolutionEngine,
) *StockHandler {
	// Initialize struct with coordinator and history engine.
	return &StockHandler{
		// Store coordinator.
		coordinator: coordinator,
		// Store history engine.
		historyEngine: historyEngine,
	}
}

// SetIntradayStore injects the optional IntradayTimelineStore dependency.
func (h *StockHandler) SetIntradayStore(store *engine.IntradayTimelineStore) {
	// Assign store pointer.
	h.intradayStore = store
}

// GetRegionalCatalog handles GET /v1/stocks/regional/catalog, returning GCC equity metadata.
func (h *StockHandler) GetRegionalCatalog(w http.ResponseWriter, r *http.Request) {
	// Set 24-hour cache control.
	w.Header().Set("Cache-Control", "public, max-age=86400, stale-while-revalidate=3600")
	// Respond with regional catalog JSON.
	RespondJSON(w, http.StatusOK, domain.GetRegionalCatalog())
}

// GetEGXStocks handles GET /v1/stocks, returning current Egyptian Stock Exchange equities.
func (h *StockHandler) GetEGXStocks(w http.ResponseWriter, r *http.Request) {
	// Serve market quotes for "egx".
	h.serveMarketStocks(w, "egx", "EGP")
}

// GetSaudiStocks handles GET /v1/stocks/saudi, returning Saudi Tadawul equity quotes.
func (h *StockHandler) GetSaudiStocks(w http.ResponseWriter, r *http.Request) {
	// Serve quotes for "saudi".
	h.serveMarketStocks(w, "saudi", "SAR")
}

// GetUAEStocks handles GET /v1/stocks/uae, returning UAE DFM and ADX equity quotes.
func (h *StockHandler) GetUAEStocks(w http.ResponseWriter, r *http.Request) {
	// Serve quotes for "uae".
	h.serveMarketStocks(w, "uae", "AED")
}

// GetQatarStocks handles GET /v1/stocks/qatar, returning Qatar QSE equity quotes.
func (h *StockHandler) GetQatarStocks(w http.ResponseWriter, r *http.Request) {
	// Serve quotes for "qatar".
	h.serveMarketStocks(w, "qatar", "QAR")
}

// GetKuwaitStocks handles GET /v1/stocks/kuwait, returning Boursa Kuwait equity quotes.
func (h *StockHandler) GetKuwaitStocks(w http.ResponseWriter, r *http.Request) {
	// Serve quotes for "kuwait".
	h.serveMarketStocks(w, "kuwait", "KWD")
}

// serveMarketStocks retrieves in-memory stock snapshot or serves a valid fallback envelope.
func (h *StockHandler) serveMarketStocks(w http.ResponseWriter, market, currency string) {
	// Retrieve typed stock snapshot from coordinator.
	snapshot, exists := h.coordinator.GetStockSnapshot(market)

	// If snapshot is missing from memory cache.
	if !exists || snapshot == nil {
		// Synthesize empty fallback stock snapshot.
		fallback := domain.StockSnapshot{
			// Data source.
			Source: "bkam-gateway",
			// Market code.
			Market: market,
			// Trading currency.
			Currency: currency,
			// Freshness flag.
			Freshness: "stale",
			// Distributed trace id.
			SourceTraceID: "cold-boot",
			// Update timestamp.
			UpdatedAt: time.Now().UTC(),
			// Scraping timestamp.
			ScrapedAt: time.Now().UTC(),
			// Empty stocks map.
			Stocks: make(map[string]domain.StockPriceItem),
		}
		// Return fallback.
		RespondJSON(w, http.StatusOK, fallback)
		// Return.
		return
	}

	// Create shallow copy to normalize envelope fields.
	res := *snapshot
	// If source is missing, set default gateway source.
	if res.Source == "" {
		// Assign gateway source.
		res.Source = "bkam-gateway"
	}
	// If currency is missing, set market currency.
	if res.Currency == "" {
		// Assign currency.
		res.Currency = currency
	}
	// Allocate container for normalized stocks.
	normalizedStocks := make(map[string]domain.StockPriceItem, len(res.Stocks))
	// Iterate through stocks map.
	for k, item := range res.Stocks {
		// If code is empty, default to dictionary key.
		if item.Code == "" {
			// Assign code.
			item.Code = k
		}
		// If symbol is empty, default to code or key.
		if item.Symbol == "" {
			// Assign symbol.
			item.Symbol = item.Code
		}
		// Store normalized stock item.
		normalizedStocks[k] = item
	}
	// Reassign normalized stocks map.
	res.Stocks = normalizedStocks

	// Transmit snapshot JSON.
	RespondJSON(w, http.StatusOK, res)
}

// GetStockHistoryByDate handles GET /v1/stocks/history/{date} single date archive queries for EGX.
func (h *StockHandler) GetStockHistoryByDate(w http.ResponseWriter, r *http.Request) {
	// Extract date.
	dateParam := strings.TrimSpace(chi.URLParam(r, "date"))

	// Validate date format.
	if _, err := time.Parse("2006-01-02", dateParam); err != nil {
		// Reject.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid date format. Expected YYYY-MM-DD.", nil)
		// Return.
		return
	}

	// Resolve from history engine.
	h.serveStockHistory(w, "egx", dateParam)
}

// isStockMarket returns true if the given code is a supported market identifier.
func isStockMarket(s string) bool {
	// Check against recognized market identifiers.
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "egx", "saudi", "uae", "qatar", "kuwait":
		// Supported market.
		return true
	default:
		// Not a market.
		return false
	}
}

// GetStockHistoryRange handles GET /v1/stocks/{symbol}/history/range and GET /v1/stocks/history/range multi-day queries.
func (h *StockHandler) GetStockHistoryRange(w http.ResponseWriter, r *http.Request) {
	// Extract raw parameter.
	rawParam := strings.TrimSpace(chi.URLParam(r, "symbol"))

	// If symbol parameter is "history" or empty, this was called as /v1/stocks/history/range.
	if rawParam == "" || strings.EqualFold(rawParam, "history") {
		// Resolve market filter with default fallback to EGX.
		market := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("market")))
		// Default market to EGX if omitted.
		if market == "" {
			market = "egx"
		}
		// Resolve optional symbol query parameter.
		sym := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
		// Serve stock range.
		h.serveStockRange(w, market, sym, r)
		// Return.
		return
	}

	// If parameter is actually a regional market name (e.g. /v1/stocks/saudi/history/range).
	if isStockMarket(rawParam) {
		// Assign canonical market identifier.
		market := strings.ToLower(rawParam)
		// Resolve optional symbol query parameter.
		sym := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
		// Serve stock range.
		h.serveStockRange(w, market, sym, r)
		// Return.
		return
	}

	// Normal single stock ticker symbol (e.g. COMI).
	symbol := strings.ToUpper(rawParam)
	// Resolve optional market query parameter.
	market := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("market")))
	// Default market to EGX.
	if market == "" {
		market = "egx"
	}
	// Serve range.
	h.serveStockRange(w, market, symbol, r)
}

// GetRegionalStockHistoryByDate handles GET /v1/stocks/{market}/history/{date} for regional GCC markets.
func (h *StockHandler) GetRegionalStockHistoryByDate(w http.ResponseWriter, r *http.Request) {
	// Extract market.
	market := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "market")))
	// Extract date.
	dateParam := strings.TrimSpace(chi.URLParam(r, "date"))

	// Validate date format.
	if _, err := time.Parse("2006-01-02", dateParam); err != nil {
		// Reject.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid date format. Expected YYYY-MM-DD.", nil)
		// Return.
		return
	}

	// Serve history.
	h.serveStockHistory(w, market, dateParam)
}

// GetRegionalStockHistoryRange handles GET /v1/stocks/{market}/{symbol}/history/range and /v1/stocks/{market}/history/range.
func (h *StockHandler) GetRegionalStockHistoryRange(w http.ResponseWriter, r *http.Request) {
	// Extract market.
	market := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "market")))
	// Extract symbol parameter.
	rawParam := strings.TrimSpace(chi.URLParam(r, "symbol"))

	// If symbol is "history" or empty, this was called as /v1/stocks/{market}/history/range.
	if rawParam == "" || strings.EqualFold(rawParam, "history") {
		// Resolve optional symbol query parameter.
		sym := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
		// Serve range.
		h.serveStockRange(w, market, sym, r)
		// Return.
		return
	}

	// Normal single symbol ticker in market.
	symbol := strings.ToUpper(rawParam)
	// Serve range.
	h.serveStockRange(w, market, symbol, r)
}

// serveStockHistory handles date resolution for stocks.
func (h *StockHandler) serveStockHistory(w http.ResponseWriter, market, date string) {
	// If history engine missing.
	if h.historyEngine == nil {
		// Return 404.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Stock history not found for %s", date), nil)
		// Return.
		return
	}

	// Resolve snapshot.
	resolved, err := h.historyEngine.GetDateSnapshot("stocks", date, domain.RangeOptions{Market: market})
	// Check error.
	if err != nil {
		// Reject.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", err.Error(), nil)
		// Return.
		return
	}

	// Check missing data.
	if resolved == nil || resolved.Data == nil {
		// Return 404.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Stock history not found for %s", date), nil)
		// Return.
		return
	}

	// Cache header.
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", resolved.CacheTTLSeconds))
	// Transmit data.
	RespondJSON(w, http.StatusOK, resolved.Data)
}

// serveStockRange handles range resolution for stocks.
func (h *StockHandler) serveStockRange(w http.ResponseWriter, market, symbol string, r *http.Request) {
	// Extract start date.
	start := strings.TrimSpace(r.URL.Query().Get("start"))
	// Extract end date.
	end := strings.TrimSpace(r.URL.Query().Get("end"))

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

	// If history engine missing, return empty list.
	if h.historyEngine == nil {
		// Return empty list.
		RespondJSON(w, http.StatusOK, []any{})
		// Return.
		return
	}

	// Query range.
	results, err := h.historyEngine.GetDateRange("stocks", domain.RangeOptions{
		StartDate: start,
		EndDate:   end,
		Market:    market,
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
		// Market exchange code.
		"market": market,
		// Starting range date.
		"startDate": start,
		// Concluding range date.
		"endDate": end,
		// Historical range items array.
		"history": results,
	}
	// If single symbol was requested, include symbol in envelope.
	if symbol != "" {
		// Security ticker symbol.
		response["symbol"] = symbol
	}
	// Transmit JSON response envelope.
	RespondJSON(w, http.StatusOK, response)
}

// GetStockIntradayLatest handles GET /v1/stocks/:symbol/intraday/latest with optional ?market= filter.
func (h *StockHandler) GetStockIntradayLatest(w http.ResponseWriter, r *http.Request) {
	// Extract symbol parameter from URL path.
	symbolParam := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "symbol")))
	// Extract optional market query parameter.
	marketParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("market")))
	// Default market to domestic EGX.
	if marketParam == "" {
		marketParam = "egx"
	}

	// Validate symbol parameter presence.
	if symbolParam == "" {
		// Reject empty symbol.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Symbol parameter is required.", nil)
		// Return early.
		return
	}

	// Default session date and freshness.
	sessionDate := time.Now().UTC().Format("2006-01-02")
	// Track whether session is today.
	isToday := true

	// If intraday store is configured.
	if h.intradayStore != nil {
		// Resolve session date.
		sessionDate, isToday = h.intradayStore.ResolveSessionDate("stocks", marketParam)
	}

	// Extract intraday points.
	intradayPoints := make([]domain.IntradayPoint, 0)
	// If intraday store available.
	if h.intradayStore != nil {
		// Retrieve points map for market.
		if _, pointsMap, exists := h.intradayStore.GetIntraday("stocks", marketParam); exists && pointsMap != nil {
			// Lookup points for requested symbol.
			intradayPoints = h.intradayStore.LookupSymbolPoints(pointsMap, symbolParam)
		}
	}

	// Price and baseline values.
	var price float64
	// Previous close baseline.
	var previousClose float64
	// Change.
	var change float64
	// Change percent.
	var changePercent float64
	// Denomination currency.
	currency := "EGP"
	// Default GCC currency if non-EGX.
	if marketParam == "saudi" {
		currency = "SAR"
	} else if marketParam == "uae" {
		currency = "AED"
	} else if marketParam == "qatar" {
		currency = "QAR"
	} else if marketParam == "kuwait" {
		currency = "KWD"
	}
	// Updated at timestamp.
	updatedAt := time.Now().UTC().Format(time.RFC3339)

	// Fetch live stock snapshot.
	snap, _ := h.coordinator.GetStockSnapshot(marketParam)
	// If snapshot exists.
	if snap != nil {
		// Update timestamp.
		updatedAt = snap.UpdatedAt.Format(time.RFC3339)
		// Set currency if present.
		if snap.Currency != "" {
			currency = snap.Currency
		}

		// Direct lookup by symbol in snapshot stocks.
		if item, found := snap.Stocks[symbolParam]; found {
			price = item.Price
			change = item.Change
			changePercent = item.ChangePercent
			if item.Currency != "" {
				currency = item.Currency
			}
			// Compute previous close baseline.
			if item.PreviousClose > 0 {
				previousClose = item.PreviousClose
			} else if price > 0 && change != 0 {
				previousClose = price - change
			} else {
				previousClose = price
			}
		} else {
			// Case-insensitive iteration across all stocks in snapshot.
			for _, item := range snap.Stocks {
				// Check symbol, code, or numeric code matches parameter.
				if strings.EqualFold(item.Symbol, symbolParam) || strings.EqualFold(item.Code, symbolParam) || strings.EqualFold(item.NumericCode, symbolParam) {
					price = item.Price
					change = item.Change
					changePercent = item.ChangePercent
					if item.Currency != "" {
						currency = item.Currency
					}
					if item.PreviousClose > 0 {
						previousClose = item.PreviousClose
					} else if price > 0 && change != 0 {
						previousClose = price - change
					} else {
						previousClose = price
					}
					break
				}
			}
		}
	}

	// Set dynamic Cache-Control header: 60s for live today, 3600s for past session.
	if isToday {
		w.Header().Set("Cache-Control", "public, max-age=60")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	// Construct response payload.
	responsePayload := domain.StockIntradayResponse{
		Symbol:        symbolParam,
		Date:          sessionDate,
		Currency:      currency,
		Price:         price,
		PreviousClose: previousClose,
		Change:        change,
		ChangePercent: changePercent,
		UpdatedAt:     updatedAt,
		Intraday:      intradayPoints,
	}

	// Transmit JSON response.
	RespondJSON(w, http.StatusOK, responsePayload)
}

