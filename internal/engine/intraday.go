// Package engine provides core domain business logic engines following Deep Module design principles.
package engine

// Import standard library packages for synchronization, temporal math, and formatting.
import (
	// fmt provides formatted string generation.
	"fmt"
	// strconv parses numeric strings.
	"strconv"
	// strings provides string manipulation and case normalization.
	"strings"
	// sync provides concurrent read-write mutex synchronization.
	"sync"
	// time provides temporal calculations and timezone conversions.
	"time"

	// domain defines intraday data structures and points.
	"bkam-api/internal/domain"
)

// IntradaySession holds a single category's intraday points map and associated calendar date.
type IntradaySession struct {
	// Date is the calendar date of the intraday points (YYYY-MM-DD).
	Date string
	// PointsMap maps asset symbols or karats to raw point arrays or structures.
	PointsMap map[string]any
	// UpdatedAt records when the session was updated in memory.
	UpdatedAt time.Time
}

// IntradayTimelineStore coordinates in-memory intraday sparkline caching and session backtracking.
type IntradayTimelineStore struct {
	// mu guards concurrent access to the session store.
	mu sync.RWMutex
	// sessions maps "category:market" keys to IntradaySession records.
	sessions map[string]*IntradaySession
	// schedule provides Cairo trading calendar and holiday checks.
	schedule *MarketScheduleEngine
	// coordinator supplies live market quotes for baseline enrichment.
	coordinator *LiveSnapshotCoordinator
	// historyEngine supplies past session history for off-hours backtracking.
	historyEngine *HistoryResolutionEngine
}

// NewIntradayTimelineStore initializes and returns a thread-safe IntradayTimelineStore instance.
func NewIntradayTimelineStore(
	schedule *MarketScheduleEngine,
	coordinator *LiveSnapshotCoordinator,
	historyEngine *HistoryResolutionEngine,
) *IntradayTimelineStore {
	// Return initialized store.
	return &IntradayTimelineStore{
		// Allocate session map.
		sessions: make(map[string]*IntradaySession),
		// Store market schedule reference.
		schedule: schedule,
		// Store live snapshot coordinator reference.
		coordinator: coordinator,
		// Store history resolution engine reference.
		historyEngine: historyEngine,
	}
}

// makeKey computes a uniform composite lookup key for category and market.
func makeKey(category, market string) string {
	// Normalize category.
	cat := strings.ToLower(strings.TrimSpace(category))
	// Normalize market.
	m := strings.ToLower(strings.TrimSpace(market))
	// If market empty, default to standard.
	if m == "" {
		// Default to primary.
		m = "default"
	}
	// Return composite key.
	return fmt.Sprintf("%s:%s", cat, m)
}

// UpdateIntraday stores updated intraday points for a category and market under a specific date.
func (s *IntradayTimelineStore) UpdateIntraday(category, market, date string, pointsMap map[string]any) {
	// Acquire exclusive write lock.
	s.mu.Lock()
	// Defer release of write lock.
	defer s.mu.Unlock()

	// Effective points dictionary.
	effectiveMap := pointsMap
	// Unwrap nested intradayPoints or points wrapper map if present.
	if inner, ok := pointsMap["intradayPoints"].(map[string]any); ok && inner != nil {
		// Use inner points map.
		effectiveMap = inner
	} else if inner, ok := pointsMap["points"].(map[string]any); ok && inner != nil {
		// Use inner points map.
		effectiveMap = inner
	}

	// Compute composite key.
	key := makeKey(category, market)
	// Store new intraday session record.
	s.sessions[key] = &IntradaySession{
		Date:      date,
		PointsMap: effectiveMap,
		UpdatedAt: time.Now().UTC(),
	}
}

// GetIntraday retrieves the in-memory intraday session for a category and market.
func (s *IntradayTimelineStore) GetIntraday(category, market string) (string, map[string]any, bool) {
	// Acquire shared read lock.
	s.mu.RLock()
	// Defer release of read lock.
	defer s.mu.RUnlock()

	// Compute composite key.
	key := makeKey(category, market)
	// Look up session.
	session, exists := s.sessions[key]
	// If absent or empty.
	if !exists || session == nil || session.PointsMap == nil {
		// Return not found.
		return "", nil, false
	}
	// Return session date and points map.
	return session.Date, session.PointsMap, true
}

// NormalizePoints transforms diverse raw point representations into a typed []domain.IntradayPoint slice.
func (s *IntradayTimelineStore) NormalizePoints(raw any) []domain.IntradayPoint {
	// If raw is nil.
	if raw == nil {
		// Return empty slice.
		return []domain.IntradayPoint{}
	}

	// Result slice.
	var result []domain.IntradayPoint

	// Check if already a slice of intraday points.
	if points, ok := raw.([]domain.IntradayPoint); ok {
		// Return directly.
		return points
	}

	// Check if raw is a slice of any.
	if list, ok := raw.([]any); ok {
		// Allocate slice with list capacity.
		result = make([]domain.IntradayPoint, 0, len(list))
		// Iterate items.
		for _, item := range list {
			// Extract map.
			if m, ok := item.(map[string]any); ok {
				// Parse point.
				p := parsePointMap(m)
				// Append.
				result = append(result, p)
			}
		}
		// Return parsed slice.
		return result
	}

	// Check if wrapped in an "elements" map.
	if wrapper, ok := raw.(map[string]any); ok {
		// If wrapper has "elements" key.
		if elements, hasElem := wrapper["elements"].([]any); hasElem {
			// Allocate slice with elements capacity.
			result = make([]domain.IntradayPoint, 0, len(elements))
			// Iterate elements.
			for _, item := range elements {
				// Extract map.
				if m, ok := item.(map[string]any); ok {
					// Parse point.
					p := parsePointMap(m)
					// Append.
					result = append(result, p)
				}
			}
			// Return parsed slice.
			return result
		}
	}

	// Return empty slice on unhandled formats.
	return []domain.IntradayPoint{}
}

// parsePointMap parses an intraday tick point map into a domain.IntradayPoint struct.
func parsePointMap(m map[string]any) domain.IntradayPoint {
	// Initialize point.
	var p domain.IntradayPoint

	// Extract time string.
	if t, ok := m["time"].(string); ok {
		// Assign time.
		p.Time = t
	}

	// Extract price.
	if priceVal, ok := m["price"]; ok {
		// Convert price.
		p.Price = toFloat64(priceVal)
	}

	// Extract optional buy.
	if buyVal, ok := m["buy"]; ok {
		// Convert buy.
		val := toFloat64(buyVal)
		// Assign pointer.
		p.Buy = &val
	}

	// Extract optional sell.
	if sellVal, ok := m["sell"]; ok {
		// Convert sell.
		val := toFloat64(sellVal)
		// Assign pointer.
		p.Sell = &val
	}

	// Return parsed point.
	return p
}

// toFloat64 safely converts a dynamic numeric value into float64.
func toFloat64(val any) float64 {
	// Type switch on numeric representation.
	switch v := val.(type) {
	case float64:
		// Return float64 directly.
		return v
	case float32:
		// Convert float32.
		return float64(v)
	case int:
		// Convert int.
		return float64(v)
	case int64:
		// Convert int64.
		return float64(v)
	case string:
		// Parse string.
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			// Return parsed float.
			return f
		}
	}
	// Default to zero.
	return 0
}

// LookupSymbolPoints performs case-resilient symbol lookup in an intraday points dictionary.
func (s *IntradayTimelineStore) LookupSymbolPoints(pointsMap map[string]any, symbol string) []domain.IntradayPoint {
	// Guard against nil map or empty symbol.
	if pointsMap == nil || strings.TrimSpace(symbol) == "" {
		// Return empty slice.
		return []domain.IntradayPoint{}
	}

	// Unwrap nested intradayPoints or points wrapper if present.
	if inner, ok := pointsMap["intradayPoints"].(map[string]any); ok && inner != nil {
		// Use inner points map.
		pointsMap = inner
	} else if inner, ok := pointsMap["points"].(map[string]any); ok && inner != nil {
		// Use inner points map.
		pointsMap = inner
	}

	// Clean target symbol.
	sym := strings.TrimSpace(symbol)
	// Build candidates slice.
	candidates := []string{
		sym,
		strings.ToUpper(sym),
		strings.ToLower(sym),
	}

	// Support karat variations (e.g. "21" vs "21k").
	if strings.HasSuffix(strings.ToLower(sym), "k") {
		// Trim suffix.
		trimmed := sym[:len(sym)-1]
		// Add candidates.
		candidates = append(candidates, trimmed, strings.ToUpper(trimmed), strings.ToLower(trimmed))
	} else {
		// Append k variants.
		candidates = append(candidates, sym+"k", sym+"K", strings.ToLower(sym)+"k", strings.ToUpper(sym)+"K")
	}

	// Support international gold aliases.
	if strings.EqualFold(sym, "international") || strings.EqualFold(sym, "intl") {
		// Add intl aliases.
		candidates = append(candidates, "international", "INTERNATIONAL", "intl", "INTL")
	}

	// Support regional Saudi numeric and alphanumeric ticker aliases.
	saudiAliases := map[string]string{
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
	// Check if symbol matches an alias.
	if saudiAlias, hasSaudiAlias := saudiAliases[strings.ToUpper(sym)]; hasSaudiAlias {
		// Add alias candidate variations.
		candidates = append(candidates, saudiAlias, strings.ToUpper(saudiAlias), strings.ToLower(saudiAlias))
	}

	// Search candidates in dictionary.
	for _, c := range candidates {
		// If candidate found in map.
		if rawPoints, found := pointsMap[c]; found && rawPoints != nil {
			// Normalize and return points.
			return s.NormalizePoints(rawPoints)
		}
	}

	// Return empty slice if not matched.
	return []domain.IntradayPoint{}
}

// ResolveSessionDate determines the active or backtracked session date for an asset class.
func (s *IntradayTimelineStore) ResolveSessionDate(category, market string) (date string, isToday bool) {
	// Compute Cairo time.
	nowCairo := s.schedule.NowCairo()
	// Current Cairo date string YYYY-MM-DD.
	todayCairo := nowCairo.Format("2006-01-02")

	// Check if in-memory store has active points for today.
	key := makeKey(category, market)
	// Acquire read lock.
	s.mu.RLock()
	// Extract session.
	session, exists := s.sessions[key]
	// Release read lock.
	s.mu.RUnlock()

	// If session exists and matches today with populated points.
	if exists && session != nil && session.Date == todayCairo && len(session.PointsMap) > 0 {
		// Active today session.
		return todayCairo, true
	}

	// If session exists and has points from earlier today or recent session.
	if exists && session != nil && len(session.PointsMap) > 0 {
		// Return cached session date.
		return session.Date, session.Date == todayCairo
	}

	// Determine market session target string for schedule engine.
	targetMarket := strings.ToUpper(market)
	// If empty or EGX.
	if targetMarket == "" || targetMarket == "DEFAULT" {
		// Match category.
		switch strings.ToLower(category) {
		case "rates":
			// Currency rates.
			targetMarket = "RATES"
		case "gold":
			// Gold commodities.
			targetMarket = "GOLD"
		default:
			// Domestic EGX stocks.
			targetMarket = "EGX"
		}
	}

	// Calculate latest trading session date via calendar engine.
	candidateDate := s.schedule.GetLatestTradingSessionDate(targetMarket, nowCairo)

	// Return resolved candidate date.
	return candidateDate, candidateDate == todayCairo
}
