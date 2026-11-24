// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for HTTP interfaces, JSON decoding, strings, and temporal math.
import (
	// encoding/json provides JSON decoding for request bodies.
	"encoding/json"
	// fmt formats error and name strings.
	"fmt"
	// math provides rounding functions.
	"math"
	// net/http provides HTTP request and response primitives.
	"net/http"
	// regexp matches identifier patterns.
	"regexp"
	// strings provides string transformations and trimming.
	"strings"
	// time provides timestamp generation.
	"time"

	// domain defines quote models and asset catalog metadata.
	"bkam-api/internal/domain"
	// engine provides the LiveSnapshotCoordinator providing in-memory quotes.
	"bkam-api/internal/engine"
)

// bankCurrencyPattern matches identifiers like CBE_USD or NBE_EUR.
var bankCurrencyPattern = regexp.MustCompile(`^[A-Za-z0-9]+_[A-Za-z]{3}$`)

// QuotesHandler coordinates multi-domain batch quote queries across currencies, gold, stocks, and indices.
type QuotesHandler struct {
	// coordinator provides zero-read in-memory snapshots.
	coordinator *engine.LiveSnapshotCoordinator
	// catalog provides bank and currency validation metadata.
	catalog *domain.CatalogRegistry
	// regionalCatalog provides GCC stock metadata.
	regionalCatalog *domain.RegionalCatalogPayload
}

// NewQuotesHandler constructs a new QuotesHandler instance.
func NewQuotesHandler(coordinator *engine.LiveSnapshotCoordinator) *QuotesHandler {
	// Return initialized handler.
	return &QuotesHandler{
		// Store coordinator.
		coordinator: coordinator,
		// Load core catalog registry.
		catalog: domain.GetCatalogRegistry(),
		// Load regional catalog payload.
		regionalCatalog: domain.GetRegionalCatalog(),
	}
}

// PostQuotes handles POST /v1/quotes batch quote queries.
func (h *QuotesHandler) PostQuotes(w http.ResponseWriter, r *http.Request) {
	// Decode JSON body.
	var req domain.BatchQuoteRequestBody
	// Parse request body.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Reject malformed body with HTTP 400.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "Invalid batch quotes request body.", nil)
		// Return early.
		return
	}

	// Validate items array presence.
	if len(req.Items) == 0 {
		// Reject empty list.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "The 'items' array must contain at least 1 item.", nil)
		// Return early.
		return
	}

	// Validate items upper limit.
	if len(req.Items) > 100 {
		// Reject oversized list.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "The 'items' array cannot exceed 100 items.", nil)
		// Return early.
		return
	}

	// Resolve batch quotes.
	h.resolveQuotes(w, req.Items)
}

// GetQuotes handles GET /v1/quotes batch quote queries with comma-separated query parameters.
func (h *QuotesHandler) GetQuotes(w http.ResponseWriter, r *http.Request) {
	// Extract items query parameter.
	rawItems := strings.TrimSpace(r.URL.Query().Get("items"))
	// Validate parameter presence.
	if rawItems == "" {
		// Reject missing parameter.
		RespondError(w, http.StatusBadRequest, "INVALID_FILTER", "The 'items' query parameter is required.", nil)
		// Return early.
		return
	}

	// Split items by comma.
	parts := strings.Split(rawItems, ",")
	// Collect non-empty items.
	items := make([]string, 0, len(parts))
	// Iterate parts.
	for _, p := range parts {
		// Trim whitespace.
		trimmed := strings.TrimSpace(p)
		// If non-empty.
		if trimmed != "" {
			// Append item.
			items = append(items, trimmed)
		}
	}

	// Validate minimum count.
	if len(items) == 0 {
		// Reject empty list.
		RespondError(w, http.StatusBadRequest, "INVALID_FILTER", "The 'items' query parameter is required.", nil)
		// Return early.
		return
	}

	// Validate upper limit.
	if len(items) > 100 {
		// Reject oversized list.
		RespondError(w, http.StatusBadRequest, "INVALID_VALIDATION", "The 'items' array cannot exceed 100 items.", nil)
		// Return early.
		return
	}

	// Resolve batch quotes.
	h.resolveQuotes(w, items)
}

// resolveQuotes resolves heterogeneous requested identifiers against in-memory snapshots.
func (h *QuotesHandler) resolveQuotes(w http.ResponseWriter, items []string) {
	// Deduplicate items while preserving order.
	seen := make(map[string]bool)
	// Unique items slice.
	uniqueItems := make([]string, 0, len(items))
	// Deduplicate loop.
	for _, it := range items {
		// Trim item.
		trimmed := strings.TrimSpace(it)
		// If valid and not seen.
		if trimmed != "" && !seen[trimmed] {
			// Mark seen.
			seen[trimmed] = true
			// Append.
			uniqueItems = append(uniqueItems, trimmed)
		}
	}

	// Initialize result maps.
	quotes := make(map[string]domain.QuoteItem)
	// Initialize unresolved slice.
	unresolved := make([]string, 0)

	// Fetch snapshots safely from in-memory coordinator.
	ratesSnap, _ := h.coordinator.GetRatesSnapshot()
	// Fetch gold snapshot.
	goldSnap, _ := h.coordinator.GetGoldSnapshot()
	// Fetch EGX stocks snapshot.
	egxStocksSnap, _ := h.coordinator.GetStockSnapshot("egx")
	// Fetch EGX indices snapshot.
	egxIndicesSnap, _ := h.coordinator.GetIndexSnapshot("egx")

	// Iterate each requested identifier.
	for _, rawID := range uniqueItems {
		// 1. Try Currency Quote.
		if q := h.resolveCurrencyQuote(rawID, ratesSnap); q != nil {
			// Store resolved quote.
			quotes[rawID] = *q
			// Continue.
			continue
		}

		// 2. Try Gold Quote.
		if q := h.resolveGoldQuote(rawID, goldSnap); q != nil {
			// Store resolved quote.
			quotes[rawID] = *q
			// Continue.
			continue
		}

		// 3. Try Index Quote.
		if q := h.resolveIndexQuote(rawID, egxIndicesSnap); q != nil {
			// Store resolved quote.
			quotes[rawID] = *q
			// Continue.
			continue
		}

		// 4. Try Stock Quote.
		if q := h.resolveStockQuote(rawID, egxStocksSnap); q != nil {
			// Store resolved quote.
			quotes[rawID] = *q
			// Continue.
			continue
		}

		// Unresolved identifier.
		unresolved = append(unresolved, rawID)
	}

	// Transmit JSON response.
	RespondJSON(w, http.StatusOK, domain.BatchQuoteResponse{
		Quotes:     quotes,
		Unresolved: unresolved,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// resolveCurrencyQuote resolves identifiers like CBE_USD against the rates snapshot.
func (h *QuotesHandler) resolveCurrencyQuote(rawID string, snap *domain.RateSnapshot) *domain.QuoteItem {
	// Check pattern match.
	if !bankCurrencyPattern.MatchString(rawID) {
		// Not a bank currency identifier.
		return nil
	}

	// If snapshot is nil.
	if snap == nil || len(snap.Rates) == 0 {
		// Snapshot unavailable.
		return nil
	}

	// Split bank code and currency code.
	parts := strings.Split(rawID, "_")
	// Uppercase bank code.
	bankCode := strings.ToUpper(parts[0])
	// Uppercase currency code.
	currencyCode := strings.ToUpper(parts[1])

	// Search rates array for match.
	for _, r := range snap.Rates {
		// Match bank code and currency code.
		if strings.EqualFold(r.BankCode, bankCode) && strings.EqualFold(r.CurrencyCode, currencyCode) {
			// Calculate change.
			var change *float64
			// Calculate change percentage.
			var changePercent *float64

			// If yesterday close rate is recorded.
			if r.YesterdayCloseSell != nil && *r.YesterdayCloseSell > 0 {
				// Compute difference.
				diff := roundTo(r.SellRate-*r.YesterdayCloseSell, 4)
				// Assign change.
				change = &diff
				// Compute percent.
				pct := roundTo((diff / *r.YesterdayCloseSell) * 100.0, 2)
				// Assign change percent.
				changePercent = &pct
			} else if r.OldSellRate != nil && *r.OldSellRate > 0 {
				// Compute fallback difference.
				diff := roundTo(r.SellRate-*r.OldSellRate, 4)
				// Assign change.
				change = &diff
				// Compute percent.
				pct := roundTo((diff / *r.OldSellRate) * 100.0, 2)
				// Assign change percent.
				changePercent = &pct
			}

			// Format human-readable display name.
			bankName := r.NameEn
			// Fallback bank name.
			if bankName == "" {
				// Use bank code.
				bankName = r.BankCode
			}
			// Composite display name.
			displayName := fmt.Sprintf("%s - %s", bankName, r.CurrencyCode)

			// Copy rates for pointers.
			buyRate := r.BuyRate
			// Sell rate.
			sellRate := r.SellRate

			// Return resolved quote.
			return &domain.QuoteItem{
				ID:            rawID,
				Category:      "currency",
				Name:          displayName,
				Price:         r.SellRate,
				BuyPrice:      &buyRate,
				SellPrice:     &sellRate,
				Change:        change,
				ChangePercent: changePercent,
				Currency:      "EGP",
				UpdatedAt:     r.RetrievedAt.Format(time.RFC3339),
			}
		}
	}

	// Match not found.
	return nil
}

// resolveGoldQuote resolves gold karat or coin identifiers (24k, 21k, 18k, pound, ounce).
func (h *QuotesHandler) resolveGoldQuote(rawID string, snap *domain.GoldSnapshot) *domain.QuoteItem {
	// Normalize key.
	key := strings.ToLower(strings.TrimSpace(rawID))
	// Valid gold keys.
	validKeys := map[string]bool{
		"24k": true, "21k": true, "18k": true, "pound": true, "ounce": true,
		"24": true, "21": true, "18": true,
	}

	// If not recognized gold key.
	if !validKeys[key] {
		// Return nil.
		return nil
	}

	// If snapshot is nil.
	if snap == nil {
		// Snapshot unavailable.
		return nil
	}

	// Normalize key without k suffix for rate lookup.
	lookupKey := key
	// If 24/21/18, add k.
	if key == "24" || key == "21" || key == "18" {
		// Standardize key.
		lookupKey = key + "k"
	}

	// Handle international gold ounce.
	if key == "ounce" {
		// Check international spot.
		if snap.International != nil && snap.International.Price > 0 {
			// Price copy.
			p := snap.International.Price
			// Return international quote.
			return &domain.QuoteItem{
				ID:            rawID,
				Category:      "gold",
				Name:          "Gold Ounce (International)",
				Price:         p,
				BuyPrice:      &p,
				SellPrice:     &p,
				Currency:      "USD",
				UpdatedAt:     snap.UpdatedAt.Format(time.RFC3339),
			}
		}
		// Unavailable.
		return nil
	}

	// Look up karat item.
	item, found := snap.Rates[lookupKey]
	// If not found.
	if !found || item.Sell == 0 {
		// Return nil.
		return nil
	}

	// Calculate change.
	var change *float64
	// Calculate change percent.
	var changePercent *float64

	// Determine reference price.
	refPrice := item.YesterdaySellPrice
	// Fallback to old sell price.
	if refPrice == 0 && item.OldSell != nil {
		// Assign old sell.
		refPrice = *item.OldSell
	}

	// If reference price exists.
	if refPrice > 0 {
		// Compute difference.
		diff := roundTo(item.Sell-refPrice, 2)
		// Assign change.
		change = &diff
		// Compute percent.
		pct := roundTo((diff/refPrice)*100.0, 2)
		// Assign percent.
		changePercent = &pct
	}

	// Determine display name.
	displayName := fmt.Sprintf("%s Gold", strings.ToUpper(lookupKey))
	// Special name for gold pound.
	if lookupKey == "pound" {
		// Set full display name.
		displayName = "Gold Pound (8g 21K)"
	}

	// Copy rates.
	buyP := item.Buy
	// Sell price.
	sellP := item.Sell

	// Return resolved gold quote.
	return &domain.QuoteItem{
		ID:            rawID,
		Category:      "gold",
		Name:          displayName,
		Price:         item.Sell,
		BuyPrice:      &buyP,
		SellPrice:     &sellP,
		Change:        change,
		ChangePercent: changePercent,
		Currency:      "EGP",
		UpdatedAt:     snap.UpdatedAt.Format(time.RFC3339),
	}
}

// resolveIndexQuote resolves benchmark index identifiers like EGX30 or TASI.
func (h *QuotesHandler) resolveIndexQuote(rawID string, snap *domain.IndexSnapshot) *domain.QuoteItem {
	// Normalize raw ID.
	upperKey := strings.ToUpper(strings.TrimSpace(rawID))

	// If EGX indices snapshot exists.
	if snap != nil && snap.Indices != nil {
		// Look up index item.
		if item, found := snap.Indices[upperKey]; found {
			// Change copy.
			ch := item.Change
			// Pct copy.
			chPct := item.ChangePercent
			// Display name.
			name := item.NameEn
			if name == "" {
				name = item.Name
			}
			if name == "" {
				name = upperKey
			}
			// Return resolved index quote.
			return &domain.QuoteItem{
				ID:            rawID,
				Category:      "index",
				Name:          name,
				Price:         item.Price,
				Change:        &ch,
				ChangePercent: &chPct,
				Currency:      "EGP",
				UpdatedAt:     snap.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	// Match not found.
	return nil
}

// resolveStockQuote resolves Egyptian or GCC equity identifiers like COMI.
func (h *QuotesHandler) resolveStockQuote(rawID string, snap *domain.StockSnapshot) *domain.QuoteItem {
	// Normalize raw ID.
	upperKey := strings.ToUpper(strings.TrimSpace(rawID))

	// Check domestic EGX stocks.
	if snap != nil && snap.Stocks != nil {
		// Helper to construct stock quote.
		makeStockQuote := func(item domain.StockPriceItem) *domain.QuoteItem {
			ch := item.Change
			chPct := item.ChangePercent
			curr := item.Currency
			if curr == "" {
				curr = snap.Currency
			}
			if curr == "" {
				curr = "EGP"
			}
			name := item.NameEn
			if name == "" {
				name = item.NameAr
			}
			if name == "" {
				name = item.Symbol
			}
			return &domain.QuoteItem{
				ID:            rawID,
				Category:      "stock",
				Name:          name,
				Price:         item.Price,
				Change:        &ch,
				ChangePercent: &chPct,
				Currency:      curr,
				UpdatedAt:     snap.UpdatedAt.Format(time.RFC3339),
			}
		}

		// Direct symbol lookup.
		if item, found := snap.Stocks[upperKey]; found {
			return makeStockQuote(item)
		}

		// Search through stocks by symbol or code.
		for _, item := range snap.Stocks {
			// Match symbol or code.
			if strings.EqualFold(item.Symbol, upperKey) || strings.EqualFold(item.Code, upperKey) {
				return makeStockQuote(item)
			}
		}
	}

	// Match not found.
	return nil
}

// roundTo rounds a float value to the specified number of decimal precision places.
func roundTo(val float64, decimals int) float64 {
	// Calculate factor 10^decimals.
	factor := math.Pow(10, float64(decimals))
	// Round and scale back.
	return math.Round(val*factor) / factor
}
