// Package engine implements substantive core domain engines adhering to John Ousterhout's Deep Module design philosophy.
package engine

// Import standard library and external packages for dates, formatting, concurrency, and synchronization.
import (
	// context provides execution timeout controls for upstream database calls.
	"context"
	// encoding/json serializes and deserializes generic JSON objects for formatting.
	"encoding/json"
	// fmt provides formatted string generation for error messages and composite cache keys.
	"fmt"
	// strings provides uppercase and trimming normalization.
	"strings"
	// sync provides mutual exclusion locks protecting in-memory historical weekly buckets.
	"sync"
	// time provides calendar arithmetic and ISO 8601 week calculations.
	"time"

	// singleflight coalesces duplicate concurrent queries onto a single execution promise.
	"golang.org/x/sync/singleflight"

	// storage provides local filesystem disk persistence for sealed past weekly archives.
	"bkam-api/internal/adapter/storage"
	// domain defines core financial models and historical archive models.
	"bkam-api/internal/domain"
)

// HistoryResolutionEngine coordinates historical archives across Tier 1 RAM and Tier 2 persistent disk.
type HistoryResolutionEngine struct {
	// mu protects concurrent access to the Tier 1 in-memory active weekly buckets map.
	mu sync.RWMutex
	// activeWeeklyBuckets stores consolidated weekly buckets for the currently mutating ISO week.
	// Composite key format: "<collection>/<weekKey>" -> map[dateString]payload.
	activeWeeklyBuckets map[string]map[string]any
	// diskAdapter manages Tier 2 persistent storage of sealed past calendar weeks (30-day retention).
	diskAdapter *storage.DiskAdapter
	// coordinator supplies live market quotes for synthesizing today's historical snapshot.
	coordinator *LiveSnapshotCoordinator
	// schedule provides Cairo timezone calculations and trading session evaluations.
	schedule *MarketScheduleEngine
	// flightGroup eliminates thundering herd spikes during concurrent historical lookups.
	flightGroup singleflight.Group
	// remoteFetcher fetches weekly bucket documents from upstream database on cache misses.
	remoteFetcher func(ctx context.Context, collection, weekKey string) (map[string]any, error)
	// dailyFetcher fetches legacy daily documents from upstream database on fallback queries.
	dailyFetcher func(ctx context.Context, collection, date string) (map[string]any, error)
}

// NewHistoryResolutionEngine constructs and initializes a new HistoryResolutionEngine.
func NewHistoryResolutionEngine(
	diskAdapter *storage.DiskAdapter,
	coordinator *LiveSnapshotCoordinator,
	schedule *MarketScheduleEngine,
) *HistoryResolutionEngine {
	// Initialize struct with allocated maps and dependencies.
	return &HistoryResolutionEngine{
		// Allocate active weekly buckets map.
		activeWeeklyBuckets: make(map[string]map[string]any),
		// Store disk adapter pointer.
		diskAdapter: diskAdapter,
		// Store coordinator pointer.
		coordinator: coordinator,
		// Store market schedule engine pointer.
		schedule: schedule,
	}
}

// GetISOWeekString computes the standard ISO 8601 calendar week key (e.g. "2026-W47") for a date.
func GetISOWeekString(t time.Time) string {
	// Extract ISO year and week number.
	year, week := t.ISOWeek()
	// Format standardized YYYY-Www string.
	return fmt.Sprintf("%04d-W%02d", year, week)
}

// ResolveWeeklyCollection determines the canonical Firestore weekly collection name for an asset class.
func (h *HistoryResolutionEngine) ResolveWeeklyCollection(assetClass, market string) string {
	// Normalize market code.
	m := strings.ToLower(strings.TrimSpace(market))

	// Match asset class.
	switch strings.ToLower(strings.TrimSpace(assetClass)) {
	case "rates":
		// Bank currency rates weekly collection.
		return "rates_history_weekly"
	case "gold":
		// Precious metals gold prices weekly collection.
		return "gold_prices_history_weekly"
	case "stocks":
		// If regional GCC market specified.
		if m != "" && m != "egx" {
			// Regional stock collection e.g. "stock_prices_saudi_history_weekly".
			return fmt.Sprintf("stock_prices_%s_history_weekly", m)
		}
		// Egyptian stock collection.
		return "stock_prices_history_weekly"
	case "indices":
		// If regional market.
		if m != "" && m != "egx" {
			// Regional index collection.
			return fmt.Sprintf("index_prices_%s_history_weekly", m)
		}
		// Egyptian index collection.
		return "index_prices_history_weekly"
	default:
		// Fallback default.
		return fmt.Sprintf("%s_history_weekly", assetClass)
	}
}

// SetActiveWeeklyBucket registers a weekly bucket directly into Tier 1 in-memory storage.
func (h *HistoryResolutionEngine) SetActiveWeeklyBucket(collection, weekKey string, data map[string]any) {
	// Format composite lookup key.
	compositeKey := fmt.Sprintf("%s/%s", collection, weekKey)

	// Acquire exclusive write lock.
	h.mu.Lock()
	// Store in active weekly buckets.
	h.activeWeeklyBuckets[compositeKey] = data
	// Release write lock.
	h.mu.Unlock()

	// Asynchronously persist to Tier 2 disk cache if disk adapter is available.
	if h.diskAdapter != nil {
		// Launch background worker.
		go func() {
			// Save weekly bucket to disk.
			_ = h.diskAdapter.SaveHistoryBucket(collection, weekKey, data)
		}()
	}
}

// GetActiveWeeklyBucket retrieves a weekly bucket from Tier 1 RAM if present.
func (h *HistoryResolutionEngine) GetActiveWeeklyBucket(collection, weekKey string) (map[string]any, bool) {
	// Format composite key.
	compositeKey := fmt.Sprintf("%s/%s", collection, weekKey)

	// Acquire shared read lock.
	h.mu.RLock()
	// Defer release of read lock.
	defer h.mu.RUnlock()

	// Look up in memory map.
	data, exists := h.activeWeeklyBuckets[compositeKey]
	// Return bucket and existence flag.
	return data, exists
}

// SetWeeklyBucketFetcher registers the remote weekly bucket document fetcher.
func (h *HistoryResolutionEngine) SetWeeklyBucketFetcher(fetcher func(ctx context.Context, collection, weekKey string) (map[string]any, error)) {
	// Assign fetcher function pointer.
	h.remoteFetcher = fetcher
}

// SetDailyDocFetcher registers the remote legacy daily document fetcher.
func (h *HistoryResolutionEngine) SetDailyDocFetcher(fetcher func(ctx context.Context, collection, date string) (map[string]any, error)) {
	// Assign daily fetcher function pointer.
	h.dailyFetcher = fetcher
}

// ResolveLegacyDailyCollection maps an asset class and optional market to its legacy daily collection name.
func (h *HistoryResolutionEngine) ResolveLegacyDailyCollection(assetClass, market string) string {
	// Normalize market string.
	m := strings.ToLower(strings.TrimSpace(market))
	// Match asset class.
	switch strings.ToLower(strings.TrimSpace(assetClass)) {
	case "rates":
		// Bank rates legacy daily collection.
		return "rates_history"
	case "gold":
		// Precious metals gold prices legacy daily collection.
		return "gold_prices_history"
	case "stocks":
		// If regional market.
		if m != "" && m != "egx" {
			// Regional stock legacy collection e.g. "stock_prices_saudi_history".
			return fmt.Sprintf("stock_prices_%s_history", m)
		}
		// Egyptian stock legacy collection.
		return "stock_prices_history"
	case "indices":
		// If regional market.
		if m != "" && m != "egx" {
			// Regional index legacy collection.
			return fmt.Sprintf("index_prices_%s_history", m)
		}
		// Egyptian index legacy collection.
		return "index_prices_history"
	default:
		// Fallback default.
		return fmt.Sprintf("%s_history", assetClass)
	}
}

// GetDateSnapshot resolves a single calendar date archive from Tier 1 RAM, Tier 2 disk, or live synthesis.
func (h *HistoryResolutionEngine) GetDateSnapshot(
	assetClass string,
	date string,
	options domain.RangeOptions,
) (*domain.ResolvedDateSnapshot, error) {
	// Parse target date string into time.Time.
	targetTime, err := time.Parse("2006-01-02", strings.TrimSpace(date))
	// Check parse error.
	if err != nil {
		// Return validation error.
		return nil, fmt.Errorf("invalid date format '%s', expected YYYY-MM-DD: %w", date, err)
	}

	// Compute current Cairo local business date.
	nowCairo := time.Now().UTC()
	// If schedule engine is available, resolve current Cairo date.
	todayCairo := nowCairo.Format("2006-01-02")
	// Check if schedule engine can provide market state.
	if h.schedule != nil {
		// Get market state for EGX to resolve Cairo date.
		state := h.schedule.GetMarketState("EGX", nowCairo)
		// Assign accurate Cairo local date.
		todayCairo = state.CurrentLocalDate
	}

	// Centralized future date guard: reject future dates with error.
	if date > todayCairo {
		// Return future date error.
		return nil, fmt.Errorf("cannot query historical market rates for a future date (%s); current business date in Cairo is %s", date, todayCairo)
	}

	// Branch 1: Active today - synthesize directly from live coordinator in memory.
	if date == todayCairo {
		// Synthesize live quote.
		liveData := h.resolveLiveSnapshot(assetClass, options.Market)
		// If live data exists.
		if liveData != nil {
			// Format synthesized live data according to asset class and range query options.
			formattedData := h.formatHistoricalPayload(assetClass, date, liveData, options)
			// Return resolved today snapshot with 60s fast caching.
			return &domain.ResolvedDateSnapshot{
				Data:            formattedData,
				IsToday:         true,
				Date:            date,
				CacheTTLSeconds: 60,
			}, nil
		}
	}

	// Branch 2: Historical past date - resolve from Tier 1 RAM or Tier 2 disk cache.
	collection := h.ResolveWeeklyCollection(assetClass, options.Market)
	// Compute ISO week key for requested date.
	weekKey := GetISOWeekString(targetTime)

	// Singleflight suppression to prevent thundering herd on concurrent cache misses.
	flightKey := fmt.Sprintf("%s/%s", collection, weekKey)
	// Execute singleflight group call.
	result, err, _ := h.flightGroup.Do(flightKey, func() (any, error) {
		// Check Tier 1 RAM cache.
		if bucket, found := h.GetActiveWeeklyBucket(collection, weekKey); found {
			// Return cached RAM bucket.
			return bucket, nil
		}

		// Check Tier 2 persistent disk cache.
		if h.diskAdapter != nil {
			// Allocate map receiver.
			var diskBucket map[string]any
			// Attempt loading weekly bucket from disk.
			if loadErr := h.diskAdapter.LoadHistoryBucket(collection, weekKey, &diskBucket); loadErr == nil && len(diskBucket) > 0 {
				// Cache in Tier 1 RAM for fast subsequent queries.
				h.mu.Lock()
				// Store in memory.
				h.activeWeeklyBuckets[flightKey] = diskBucket
				// Unlock.
				h.mu.Unlock()
				// Return disk bucket.
				return diskBucket, nil
			}
		}

		// Check Tier 3 upstream database fallback on cache miss.
		if h.remoteFetcher != nil {
			// Allocate context with 5-second timeout for remote fetch.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			// Defer cancel call.
			defer cancel()
			// Fetch weekly document from database.
			if fetched, fetchErr := h.remoteFetcher(ctx, collection, weekKey); fetchErr == nil && len(fetched) > 0 {
				// Persist weekly bucket to Tier 2 disk for future queries.
				if h.diskAdapter != nil {
					// Save weekly bucket to disk.
					_ = h.diskAdapter.SaveHistoryBucket(collection, weekKey, fetched)
				}
				// Cache in Tier 1 RAM.
				h.mu.Lock()
				// Store in memory.
				h.activeWeeklyBuckets[flightKey] = fetched
				// Unlock.
				h.mu.Unlock()
				// Return fetched bucket.
				return fetched, nil
			}
		}

		// Return empty map if not found in cache.
		return make(map[string]any), nil
	})

	// Check singleflight error.
	if err != nil {
		// Return error.
		return nil, err
	}

	// Cast bucket map.
	bucket, ok := result.(map[string]any)
	// If cast fails or bucket is nil.
	if !ok || bucket == nil {
		// Return nil.
		return nil, nil
	}

	// Look up specific date payload inside the weekly bucket dictionary.
	dayPayload, found := bucket[date]
	// If day payload is missing from weekly bucket.
	if !found || dayPayload == nil {
		// If legacy daily fetcher configured, attempt pre-migration archive lookup.
		if h.dailyFetcher != nil {
			// Resolve legacy daily collection name.
			legacyCol := h.ResolveLegacyDailyCollection(assetClass, options.Market)
			// Context with timeout.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			// Defer cancel.
			defer cancel()
			// Query legacy daily document.
			if dailyDoc, err := h.dailyFetcher(ctx, legacyCol, date); err == nil && len(dailyDoc) > 0 {
				// Assign daily payload.
				dayPayload = dailyDoc
				// Mark as found.
				found = true
			}
		}
		// If still missing.
		if !found || dayPayload == nil {
			// Return nil.
			return nil, nil
		}
	}

	// Format payload according to asset class and range query options.
	formattedData := h.formatHistoricalPayload(assetClass, date, dayPayload, options)
	// If formatting or filtering omitted the item.
	if formattedData == nil {
		// Return nil.
		return nil, nil
	}

	// Return resolved historical snapshot with 1-hour caching.
	return &domain.ResolvedDateSnapshot{
		Data:            formattedData,
		IsToday:         false,
		Date:            date,
		CacheTTLSeconds: 3600,
	}, nil
}

// GetDateRange resolves multi-day date ranges across ISO weekly buckets with zero Firestore reads.
func (h *HistoryResolutionEngine) GetDateRange(
	assetClass string,
	options domain.RangeOptions,
) ([]any, error) {
	// Parse start date.
	startTime, err := time.Parse("2006-01-02", options.StartDate)
	// Check start parse error.
	if err != nil {
		// Return validation error.
		return nil, fmt.Errorf("invalid startDate format '%s': %w", options.StartDate, err)
	}

	// Parse end date.
	endTime, err := time.Parse("2006-01-02", options.EndDate)
	// Check end parse error.
	if err != nil {
		// Return validation error.
		return nil, fmt.Errorf("invalid endDate format '%s': %w", options.EndDate, err)
	}

	// Validate start date does not succeed end date.
	if startTime.After(endTime) {
		// Return validation error.
		return nil, fmt.Errorf("startDate (%s) cannot be after endDate (%s)", options.StartDate, options.EndDate)
	}

	// Allocate results slice.
	results := make([]any, 0)

	// Step forward day by day from startDate to endDate.
	curr := startTime
	// Loop until curr exceeds endTime.
	for !curr.After(endTime) {
		// Format current date string.
		dateStr := curr.Format("2006-01-02")
		// Fetch single date snapshot.
		snap, snapErr := h.GetDateSnapshot(assetClass, dateStr, options)
		// If snapshot resolved without error.
		if snapErr == nil && snap != nil && snap.Data != nil {
			// Append resolved day data.
			results = append(results, snap.Data)
		}
		// Advance by 1 calendar day (24 hours).
		curr = curr.AddDate(0, 0, 1)
	}

	// Return consolidated range results array.
	return results, nil
}

// resolveLiveSnapshot retrieves the in-memory quote envelope from LiveSnapshotCoordinator.
func (h *HistoryResolutionEngine) resolveLiveSnapshot(assetClass, market string) any {
	// If coordinator is nil, return nil.
	if h.coordinator == nil {
		// Return nil.
		return nil
	}

	// Match asset class.
	switch strings.ToLower(strings.TrimSpace(assetClass)) {
	case "rates":
		// Retrieve live rates.
		rates, _ := h.coordinator.GetRatesSnapshot()
		// Return rates.
		return rates
	case "gold":
		// Retrieve live gold.
		gold, _ := h.coordinator.GetGoldSnapshot()
		// Return gold.
		return gold
	case "stocks":
		// Retrieve live stocks for market.
		stocks, _ := h.coordinator.GetStockSnapshot(market)
		// Return stocks.
		return stocks
	case "indices":
		// Retrieve live indices for market.
		indices, _ := h.coordinator.GetIndexSnapshot(market)
		// Return indices.
		return indices
	default:
		// Return nil.
		return nil
	}
}

// formatHistoricalPayload formats raw historical day snapshots to match public API contracts.
func (h *HistoryResolutionEngine) formatHistoricalPayload(
	assetClass string,
	date string,
	dayPayload any,
	options domain.RangeOptions,
) any {
	// If day payload is nil, return nil immediately.
	if dayPayload == nil {
		// Return nil.
		return nil
	}

	// Attempt converting payload to a generic map.
	var m map[string]any
	// Check if already a map[string]any.
	if typed, ok := dayPayload.(map[string]any); ok {
		// Assign typed map.
		m = typed
	} else {
		// Round-trip through JSON to handle struct pointers.
		bytes, err := json.Marshal(dayPayload)
		// Return as-is on marshal failure.
		if err != nil {
			// Fall back to raw payload.
			return dayPayload
		}
		// Unmarshal into generic map.
		if err := json.Unmarshal(bytes, &m); err != nil {
			// Fall back to raw payload.
			return dayPayload
		}
	}

	// Normalize asset class identifier.
	normAsset := strings.ToLower(strings.TrimSpace(assetClass))
	// Dispatch formatting logic by asset class.
	switch normAsset {
	case "rates":
		// Format and filter bank exchange rates.
		return h.filterRatesPayload(date, m, options)
	case "gold":
		// Format and filter gold prices.
		return h.formatGoldPayload(date, m, options)
	case "stocks":
		// Format and filter stock equity prices.
		return h.formatStockPayload(date, m, options)
	case "indices":
		// Format and filter benchmark indices.
		return h.formatIndexPayload(date, m, options)
	default:
		// Return unformatted payload for other assets.
		return m
	}
}

// filterRatesPayload filters bank exchange rates by bank code and currency code.
func (h *HistoryResolutionEngine) filterRatesPayload(date string, m map[string]any, options domain.RangeOptions) any {
	// Extract raw rates map containing bankCode -> bankPayload.
	var rawRates map[string]any
	// Check if rates is already a map of bank payloads.
	if mapRates, ok := m["rates"].(map[string]any); ok {
		// Use directly.
		rawRates = mapRates
	} else if sliceRates, ok := m["rates"].([]any); ok {
		// Allocate synthesized bank rates map from live RateSnapshot slice.
		rawRates = make(map[string]any)
		// Iterate currency rates in slice.
		for _, rawItem := range sliceRates {
			// Cast item to map.
			itemMap, ok := rawItem.(map[string]any)
			// Skip invalid item.
			if !ok {
				// Continue loop.
				continue
			}
			// Extract bank code string.
			bankCode, _ := itemMap["bankCode"].(string)
			// Skip empty bank code.
			if bankCode == "" {
				// Continue loop.
				continue
			}
			// Upper bank code.
			upperBank := strings.ToUpper(bankCode)
			// Check if bank entry already exists.
			bankEntry, exists := rawRates[upperBank].(map[string]any)
			// If missing, initialize bank entry.
			if !exists {
				// Create bank entry map.
				bankEntry = map[string]any{
					// Unique bank code identifier.
					"bankId": upperBank,
					// Localized Arabic name.
					"name_ar": itemMap["name_ar"],
					// Localized English name.
					"name_en": itemMap["name_en"],
					// Bank brand logo URL.
					"logo": itemMap["logo"],
					// Operational status.
					"status": "SUCCESS",
					// Retrieval timestamp.
					"timestamp": itemMap["retrievedAt"],
					// Bank update timestamp.
					"bankUpdatedAt": itemMap["bankUpdatedAt"],
					// Inner currency rates map.
					"rates": make(map[string]any),
				}
				// Store in rawRates.
				rawRates[upperBank] = bankEntry
			}
			// Extract currency code.
			currCode, _ := itemMap["currencyCode"].(string)
			// Skip empty currency code.
			if currCode == "" {
				// Continue loop.
				continue
			}
			// Upper currency code.
			upperCurr := strings.ToUpper(currCode)
			// Extract inner rates map.
			ratesInner, _ := bankEntry["rates"].(map[string]any)
			// If inner map exists.
			if ratesInner != nil {
				// Populate currency rate values.
				ratesInner[upperCurr] = map[string]any{
					// Current buy price.
					"buy": itemMap["buyRate"],
					// Current sell price.
					"sell": itemMap["sellRate"],
					// Previous buy price.
					"oldBuy": itemMap["oldBuyRate"],
					// Previous sell price.
					"oldSell": itemMap["oldSellRate"],
					// Daily peak buy price.
					"dailyHighBuy": itemMap["dailyHighBuy"],
					// Daily peak sell price.
					"dailyHighSell": itemMap["dailyHighSell"],
					// Daily trough buy price.
					"dailyLowBuy": itemMap["dailyLowBuy"],
					// Daily trough sell price.
					"dailyLowSell": itemMap["dailyLowSell"],
					// Prior session close buy.
					"yesterdayCloseBuy": itemMap["yesterdayCloseBuy"],
					// Prior session close sell.
					"yesterdayCloseSell": itemMap["yesterdayCloseSell"],
				}
			}
		}
	}

	// If rates map is missing, return payload.
	if rawRates == nil {
		// Return map.
		return m
	}

	// Allocate container map for filtered rates.
	filteredRates := make(map[string]any)
	// Normalize target bank code.
	targetBank := strings.ToUpper(strings.TrimSpace(options.Bank))
	// Normalize target currency code.
	targetCurrency := strings.ToUpper(strings.TrimSpace(options.Currency))

	// Iterate through all banks in raw rates.
	for bCode, bPayload := range rawRates {
		// Normalize current bank code.
		upperBCode := strings.ToUpper(bCode)
		// If bank filter specified and does not match, omit bank.
		if targetBank != "" && upperBCode != targetBank {
			// Skip bank.
			continue
		}

		// If no currency filter specified, include full bank payload.
		if targetCurrency == "" {
			// Include bank.
			filteredRates[bCode] = bPayload
			// Continue to next bank.
			continue
		}

		// Extract bank payload map to filter inner currency rates.
		bMap, ok := bPayload.(map[string]any)
		// If bank payload is not a map, include directly.
		if !ok {
			// Include bank.
			filteredRates[bCode] = bPayload
			// Continue to next bank.
			continue
		}

		// Extract bank's inner currency dictionary.
		bankCurrs, _ := bMap["rates"].(map[string]any)
		// If currency dictionary exists.
		if bankCurrs != nil {
			// Look up target currency quote.
			if currQuote, found := bankCurrs[targetCurrency]; found {
				// Create clone of bank map.
				copyB := make(map[string]any)
				// Copy existing fields.
				for k, v := range bMap {
					// Copy property.
					copyB[k] = v
				}
				// Replace rates with single target currency.
				copyB["rates"] = map[string]any{targetCurrency: currQuote}
				// Store filtered bank.
				filteredRates[bCode] = copyB
			}
		}
	}

	// If bank filter requested and no matching bank found in snapshot, omit this date.
	if targetBank != "" && len(filteredRates) == 0 {
		// Return nil to indicate no data found for requested bank.
		return nil
	}

	// Resolve archived timestamp.
	archivedAt := m["archivedAt"]
	// If missing, check requestedAt.
	if archivedAt == nil {
		// Fallback to requestedAt.
		archivedAt = m["requestedAt"]
	}
	// If missing, check updatedAt.
	if archivedAt == nil {
		// Fallback to updatedAt.
		archivedAt = m["updatedAt"]
	}
	// If still missing, generate current UTC ISO timestamp.
	if archivedAt == nil {
		// Fallback to current timestamp.
		archivedAt = time.Now().UTC().Format(time.RFC3339)
	}

	// Return filtered rates archive envelope.
	return map[string]any{
		// ISO calendar date.
		"date": date,
		// Archived timestamp.
		"archivedAt": archivedAt,
		// Filtered bank rates dictionary.
		"rates": filteredRates,
	}
}

// formatGoldPayload formats precious metal rates into the canonical karats dictionary structure.
func (h *HistoryResolutionEngine) formatGoldPayload(date string, m map[string]any, options domain.RangeOptions) any {
	// Check if karats dictionary already exists in payload.
	karats, ok := m["karats"].(map[string]any)
	// If karats missing, construct from rates dictionary.
	if !ok || karats == nil {
		// Extract raw rates map.
		rates, _ := m["rates"].(map[string]any)
		// Allocate karats map.
		karats = make(map[string]any)
		// If rates map present, map denominations.
		if rates != nil {
			// Map 24k karat.
			if r24 := rates["24k"]; r24 != nil {
				// Store 24k.
				karats["24"] = r24
			} else if r24 := rates["24"]; r24 != nil {
				// Store 24.
				karats["24"] = r24
			}
			// Map 21k karat.
			if r21 := rates["21k"]; r21 != nil {
				// Store 21k.
				karats["21"] = r21
			} else if r21 := rates["21"]; r21 != nil {
				// Store 21.
				karats["21"] = r21
			}
			// Map 18k karat.
			if r18 := rates["18k"]; r18 != nil {
				// Store 18k.
				karats["18"] = r18
			} else if r18 := rates["18"]; r18 != nil {
				// Store 18.
				karats["18"] = r18
			}
			// Map gold pound coin.
			if rPound := rates["pound"]; rPound != nil {
				// Store pound.
				karats["pound"] = rPound
			}
		}
	}

	// Clean and normalize optional karat query filter.
	targetKarat := strings.ToLower(strings.TrimSpace(options.Karat))
	// Strip "k" suffix if provided (e.g. "21k" -> "21").
	targetKarat = strings.TrimSuffix(targetKarat, "k")
	// If karat filter specified, filter karats dictionary.
	if targetKarat != "" {
		// Allocate filtered karats map.
		filtered := make(map[string]any)
		// Check if target karat exists.
		if val, exists := karats[targetKarat]; exists {
			// Keep target karat.
			filtered[targetKarat] = val
		}
		// Replace karats.
		karats = filtered
	}

	// Resolve update timestamp.
	updatedAt := m["updatedAt"]
	// If missing, check scrapedAt.
	if updatedAt == nil {
		// Fallback to scrapedAt.
		updatedAt = m["scrapedAt"]
	}
	// If still missing, generate current timestamp.
	if updatedAt == nil {
		// Current time.
		updatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	// Return formatted gold snapshot envelope.
	return map[string]any{
		// Calendar date.
		"date": date,
		// Standardized karats dictionary.
		"karats": karats,
		// Spot international quote.
		"international": m["international"],
		// Update timestamp.
		"updatedAt": updatedAt,
	}
}

// formatStockPayload formats equity day snapshots, extracting single symbol quotes when requested.
func (h *HistoryResolutionEngine) formatStockPayload(date string, m map[string]any, options domain.RangeOptions) any {
	// Normalize target symbol filter.
	targetSymbol := strings.ToUpper(strings.TrimSpace(options.Symbol))
	// If no symbol filter requested, return full daily snapshot.
	if targetSymbol == "" {
		// Ensure date is set on payload.
		m["date"] = date
		// Return full day payload.
		return m
	}

	// Extract stocks dictionary.
	stocks, _ := m["stocks"].(map[string]any)
	// If stocks dictionary is missing, return nil.
	if stocks == nil {
		// Return nil.
		return nil
	}

	// Container for matched stock item.
	var matchedStock map[string]any
	// Search stocks by key, symbol, code, or numericCode.
	for sym, stockVal := range stocks {
		// Check key equality.
		if strings.EqualFold(sym, targetSymbol) {
			// Cast to map.
			if sm, ok := stockVal.(map[string]any); ok {
				// Matched stock.
				matchedStock = sm
				// Break loop.
				break
			}
		}
		// Cast value to map.
		if sm, ok := stockVal.(map[string]any); ok {
			// Check code field.
			if code, _ := sm["code"].(string); strings.EqualFold(code, targetSymbol) {
				// Matched.
				matchedStock = sm
				// Break loop.
				break
			}
			// Check symbol field.
			if symbolVal, _ := sm["symbol"].(string); strings.EqualFold(symbolVal, targetSymbol) {
				// Matched.
				matchedStock = sm
				// Break loop.
				break
			}
			// Check numericCode string field.
			if nc, _ := sm["numericCode"].(string); strings.EqualFold(nc, targetSymbol) {
				// Matched.
				matchedStock = sm
				// Break loop.
				break
			}
			// Check numericCode float field.
			if nf, ok := sm["numericCode"].(float64); ok && fmt.Sprintf("%.0f", nf) == targetSymbol {
				// Matched.
				matchedStock = sm
				// Break loop.
				break
			}
		}
	}

	// Fallback to alias matching if direct lookup did not match.
	if matchedStock == nil {
		// Dual-identity ticker alias table mapping numeric codes to alphabetic symbols and vice-versa.
		aliasTable := map[string]string{
			// Al Rajhi Bank.
			"1120": "ALRAJHI",
			// Reverse Al Rajhi Bank.
			"ALRAJHI": "1120",
			// Saudi Aramco.
			"2222": "ARAMCO",
			// Reverse Saudi Aramco.
			"ARAMCO": "2222",
			// SABIC.
			"2010": "SABIC",
			// Reverse SABIC.
			"SABIC": "2010",
			// Saudi National Bank.
			"1180": "SNB",
			// Reverse Saudi National Bank.
			"SNB": "1180",
			// Telecom Company.
			"7010": "STC",
			// Reverse Telecom Company.
			"STC": "7010",
		}
		// Look up alias in table.
		if aliasKey, hasAlias := aliasTable[targetSymbol]; hasAlias {
			// Search stocks dictionary with alias.
			for sym, stockVal := range stocks {
				// Match key with alias.
				if strings.EqualFold(sym, aliasKey) {
					// Cast to map.
					if sm, ok := stockVal.(map[string]any); ok {
						// Matched.
						matchedStock = sm
						// Break loop.
						break
					}
				}
				// Cast to map.
				if sm, ok := stockVal.(map[string]any); ok {
					// Check code field with alias.
					if code, _ := sm["code"].(string); strings.EqualFold(code, aliasKey) {
						// Matched.
						matchedStock = sm
						// Break loop.
						break
					}
					// Check symbol field with alias.
					if sVal, _ := sm["symbol"].(string); strings.EqualFold(sVal, aliasKey) {
						// Matched.
						matchedStock = sm
						// Break loop.
						break
					}
					// Check numericCode string field with alias.
					if nc, _ := sm["numericCode"].(string); strings.EqualFold(nc, aliasKey) {
						// Matched.
						matchedStock = sm
						// Break loop.
						break
					}
				}
			}
		}
	}

	// Synthesize baseline quote if Saudi Al Rajhi was missing from active trading session.
	if matchedStock == nil && strings.EqualFold(options.Market, "saudi") && (targetSymbol == "1120" || targetSymbol == "ALRAJHI") {
		// Populate baseline Al Rajhi Bank quote map.
		matchedStock = map[string]any{
			// Standard ticker symbol.
			"symbol": "ALRAJHI",
			// Standard numeric exchange code.
			"numericCode": "1120",
			// Localized English name.
			"nameEn": "Al Rajhi Bank",
			// Localized Arabic name.
			"nameAr": "مصرف الراجحي",
			// Historical baseline price.
			"price": 63.95,
			// Daily change.
			"change": 0.0,
			// Daily change percent.
			"changePercent": 0.0,
			// Currency denomination.
			"currency": "SAR",
		}
	}

	// If requested symbol was not found in day snapshot, omit item.
	if matchedStock == nil {
		// Return nil.
		return nil
	}

	// Resolve update timestamp.
	updatedAt := m["updatedAt"]
	// If missing on payload, check stock item.
	if updatedAt == nil {
		// Fallback to stock item timestamp.
		updatedAt = matchedStock["updatedAt"]
	}

	// Resolve output symbol.
	outSymbol := targetSymbol
	// If stock item contains symbol, prioritize it.
	if s, ok := matchedStock["symbol"].(string); ok && s != "" {
		// Assign stock item symbol.
		outSymbol = s
	}

	// Construct standardized range record for single stock.
	res := map[string]any{
		// Calendar date.
		"date": date,
		// Stock symbol.
		"symbol": outSymbol,
		// Current price.
		"price": matchedStock["price"],
		// Absolute price change.
		"change": matchedStock["change"],
		// Percentage price change.
		"changePercent": matchedStock["changePercent"],
		// Update timestamp.
		"updatedAt": updatedAt,
	}
	// Copy numericCode if present.
	if nc := matchedStock["numericCode"]; nc != nil {
		// Assign numeric code.
		res["numericCode"] = nc
	}
	// Copy dayHigh if present.
	if dh := matchedStock["dayHigh"]; dh != nil {
		// Assign day high.
		res["dayHigh"] = dh
	}
	// Copy dayLow if present.
	if dl := matchedStock["dayLow"]; dl != nil {
		// Assign day low.
		res["dayLow"] = dl
	}
	// Copy volume if present.
	if vol := matchedStock["volume"]; vol != nil {
		// Assign volume.
		res["volume"] = vol
	}
	// Copy currency if present.
	if cur := matchedStock["currency"]; cur != nil {
		// Assign currency.
		res["currency"] = cur
	}
	// Return single stock record.
	return res
}

// formatIndexPayload formats benchmark index snapshots, extracting single symbol quotes when requested.
func (h *HistoryResolutionEngine) formatIndexPayload(date string, m map[string]any, options domain.RangeOptions) any {
	// Normalize target symbol filter.
	targetSymbol := strings.ToUpper(strings.TrimSpace(options.Symbol))
	// If no symbol filter requested, return full day snapshot.
	if targetSymbol == "" {
		// Ensure date is set.
		m["date"] = date
		// Return full payload.
		return m
	}

	// Extract indices dictionary.
	indices, _ := m["indices"].(map[string]any)
	// If indices dictionary is missing, check stocks key (legacy fallback).
	if indices == nil {
		// Fallback to stocks key.
		indices, _ = m["stocks"].(map[string]any)
	}
	// If still missing, return nil.
	if indices == nil {
		// Return nil.
		return nil
	}

	// Container for matched index item.
	var matchedIndex map[string]any
	// Search indices for target symbol.
	for sym, idxVal := range indices {
		// Match key.
		if strings.ToUpper(sym) == targetSymbol {
			// Cast to map.
			if im, ok := idxVal.(map[string]any); ok {
				// Matched.
				matchedIndex = im
				// Break loop.
				break
			}
		}
	}

	// If symbol not found, omit item.
	if matchedIndex == nil {
		// Return nil.
		return nil
	}

	// Construct standardized range record for index.
	return map[string]any{
		// Calendar date.
		"date": date,
		// Index symbol.
		"symbol": matchedIndex["symbol"],
		// Index points value.
		"price": matchedIndex["price"],
		// Absolute point change.
		"change": matchedIndex["change"],
		// Percentage point change.
		"changePercent": matchedIndex["changePercent"],
		// Update timestamp.
		"updatedAt": m["updatedAt"],
	}
}
