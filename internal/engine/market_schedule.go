// Package engine implements substantive core domain engines adhering to John Ousterhout's Deep Module design philosophy.
package engine

// Import standard library packages for time calculation, timezone loading, string manipulation, and synchronization.
import (
	// strings provides string normalization and uppercase conversion.
	"strings"
	// sync provides thread-safe access to engine holiday maps and caches.
	"sync"
	// time provides temporal calculations and timezone location models.
	"time"
	// time/tzdata embeds the official IANA timezone database directly into the compiled executable binary.
	_ "time/tzdata"
)

// MarketID represents canonical exchange market codes across domestic and GCC regions.
type MarketID string

// Canonical market identifier constants.
const (
	// MarketEGX represents the Egyptian Stock Exchange equities market.
	MarketEGX MarketID = "EGX"
	// MarketEGXIndices represents Egyptian benchmark indices (EGX30, EGX70, EGX100).
	MarketEGXIndices MarketID = "EGX_INDICES"
	// MarketSA represents the Saudi Stock Exchange (Tadawul).
	MarketSA MarketID = "SA"
	// MarketSAIndices represents Saudi Tadawul index benchmarks (TASI).
	MarketSAIndices MarketID = "SA_INDICES"
	// MarketAE represents the United Arab Emirates financial markets (DFM and ADX).
	MarketAE MarketID = "AE"
	// MarketAEIndices represents UAE index benchmarks (DFMGI, FADGI).
	MarketAEIndices MarketID = "AE_INDICES"
	// MarketQA represents the Qatar Stock Exchange (QSE).
	MarketQA MarketID = "QA"
	// MarketQAIndices represents Qatar benchmark indices.
	MarketQAIndices MarketID = "QA_INDICES"
	// MarketKW represents Boursa Kuwait (KSE).
	MarketKW MarketID = "KW"
	// MarketKWIndices represents Kuwait benchmark indices.
	MarketKWIndices MarketID = "KW_INDICES"
	// MarketRates represents Central Bank of Egypt and commercial bank FX rates.
	MarketRates MarketID = "RATES"
	// MarketGold represents Egyptian precious metals (gold karats, gold pound, ounce).
	MarketGold MarketID = "GOLD"
	// MarketCommodities represents domestic fuel petroleum and commodities.
	MarketCommodities MarketID = "COMMODITIES"
)

// MarketScheduleConfig defines the operational trading hours, timezone, and active weekdays for an exchange.
type MarketScheduleConfig struct {
	// ID is the canonical market identifier.
	ID MarketID
	// Timezone is the IANA timezone string (e.g. "Africa/Cairo", "Asia/Riyadh").
	Timezone string
	// Location is the cached *time.Location loaded for this market.
	Location *time.Location
	// ActiveDays lists integer weekdays (0 = Sunday, 1 = Monday, ..., 6 = Saturday) when trading occurs.
	ActiveDays []int
	// OpenHour is the local hour when the trading window opens (24-hour clock).
	OpenHour int
	// OpenMinute is the local minute when the trading window opens.
	OpenMinute int
	// CloseHour is the local hour when the trading window closes (24-hour clock).
	CloseHour int
	// CloseMinute is the local minute when the trading window closes.
	CloseMinute int
	// Aliases lists alternative lookup names for this market (e.g. "TADAWUL", "EGYPT").
	Aliases []string
}

// MarketState reflects comprehensive trading session status, next opening bell, and cache TTL offsets.
type MarketState struct {
	// Market identifies the evaluated exchange.
	Market MarketID `json:"market"`
	// CurrentLocalDate represents the current calendar date in the exchange's timezone (YYYY-MM-DD).
	CurrentLocalDate string `json:"currentLocalDate"`
	// LocalDayOfWeek represents the integer weekday index (0=Sunday through 6=Saturday).
	LocalDayOfWeek int `json:"localDayOfWeek"`
	// LocalHourDecimal represents the local time as a fractional decimal hour (e.g. 10.5 for 10:30).
	LocalHourDecimal float64 `json:"localHourDecimal"`
	// IsOpen indicates whether the exchange is currently actively trading.
	IsOpen bool `json:"isOpen"`
	// NextOpenTimestampMs indicates the Unix millisecond timestamp of the next opening bell.
	NextOpenTimestampMs int64 `json:"nextOpenTimestampMs"`
	// SecondsUntilNextOpen indicates duration in seconds until the next opening session.
	SecondsUntilNextOpen int `json:"secondsUntilNextOpen"`
	// LatestTradingSessionDate specifies the YYYY-MM-DD date of the most recent completed or active session.
	LatestTradingSessionDate string `json:"latestTradingSessionDate"`
}

// MarketScheduleEngine coordinates trading session hours, holiday closures, and dynamic cache TTL calculations.
type MarketScheduleEngine struct {
	// mu protects concurrent modifications to custom holiday sets.
	mu sync.RWMutex
	// configs maps canonical MarketID to MarketScheduleConfig.
	configs map[MarketID]*MarketScheduleConfig
	// aliasLookup maps uppercase alias strings to canonical MarketID.
	aliasLookup map[string]MarketID
	// holidays maps MarketID to sets of holiday date strings (YYYY-MM-DD).
	holidays map[MarketID]map[string]bool
	// cairoLoc points to the cached *time.Location for Africa/Cairo.
	cairoLoc *time.Location
}

// NewMarketScheduleEngine constructs and returns a fully initialized MarketScheduleEngine.
func NewMarketScheduleEngine() *MarketScheduleEngine {
	// Load Africa/Cairo timezone location.
	cairoLoc, err := time.LoadLocation("Africa/Cairo")
	// If loading fails, fallback to UTC to prevent nil pointer crashes.
	if err != nil {
		// Use UTC fallback.
		cairoLoc = time.UTC
	}

	// Load Riyadh timezone location.
	riyadhLoc, err := time.LoadLocation("Asia/Riyadh")
	// Fallback check.
	if err != nil {
		// Fallback to UTC.
		riyadhLoc = time.UTC
	}

	// Load Dubai timezone location.
	dubaiLoc, err := time.LoadLocation("Asia/Dubai")
	// Fallback check.
	if err != nil {
		// Fallback to UTC.
		dubaiLoc = time.UTC
	}

	// Load Qatar timezone location.
	qatarLoc, err := time.LoadLocation("Asia/Qatar")
	// Fallback check.
	if err != nil {
		// Fallback to UTC.
		qatarLoc = time.UTC
	}

	// Load Kuwait timezone location.
	kuwaitLoc, err := time.LoadLocation("Asia/Kuwait")
	// Fallback check.
	if err != nil {
		// Fallback to UTC.
		kuwaitLoc = time.UTC
	}

	// Initialize engine struct.
	engine := &MarketScheduleEngine{
		// Allocate configs map.
		configs: make(map[MarketID]*MarketScheduleConfig),
		// Allocate alias lookup map.
		aliasLookup: make(map[string]MarketID),
		// Allocate holidays map.
		holidays: make(map[MarketID]map[string]bool),
		// Store Cairo timezone location.
		cairoLoc: cairoLoc,
	}

	// Define raw schedule definitions.
	rawConfigs := []*MarketScheduleConfig{
		// Egyptian Stock Exchange Equities: Buffered 09:00 - 15:30 Cairo time (Sunday - Thursday).
		{
			ID:          MarketEGX,
			Timezone:    "Africa/Cairo",
			Location:    cairoLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    9,
			OpenMinute:  0,
			CloseHour:   15,
			CloseMinute: 30,
			Aliases:     []string{"EG", "EGYPT", "STOCKS", "STOCK"},
		},
		// Egyptian Stock Exchange Indices: Aligned with EGX equities session.
		{
			ID:          MarketEGXIndices,
			Timezone:    "Africa/Cairo",
			Location:    cairoLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    9,
			OpenMinute:  0,
			CloseHour:   15,
			CloseMinute: 30,
			Aliases:     []string{"INDICES", "INDEX", "EGX30", "EGX70", "EGX100"},
		},
		// Saudi Stock Exchange (Tadawul): Buffered 09:00 - 16:00 Riyadh time (Sunday - Thursday).
		{
			ID:          MarketSA,
			Timezone:    "Asia/Riyadh",
			Location:    riyadhLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    9,
			OpenMinute:  0,
			CloseHour:   16,
			CloseMinute: 0,
			Aliases:     []string{"KSA", "SAUDI", "TADAWUL"},
		},
		// Saudi Tadawul Indices: Aligned with Saudi main session.
		{
			ID:          MarketSAIndices,
			Timezone:    "Asia/Riyadh",
			Location:    riyadhLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    9,
			OpenMinute:  0,
			CloseHour:   16,
			CloseMinute: 0,
			Aliases:     []string{"TASI", "MT30"},
		},
		// UAE Financial Markets (DFM & ADX): Buffered 09:00 - 16:00 Dubai time (Monday - Friday).
		{
			ID:          MarketAE,
			Timezone:    "Asia/Dubai",
			Location:    dubaiLoc,
			ActiveDays:  []int{1, 2, 3, 4, 5},
			OpenHour:    9,
			OpenMinute:  0,
			CloseHour:   16,
			CloseMinute: 0,
			Aliases:     []string{"UAE", "DUBAI", "DFM", "ADX"},
		},
		// UAE Indices: Aligned with UAE main session.
		{
			ID:          MarketAEIndices,
			Timezone:    "Asia/Dubai",
			Location:    dubaiLoc,
			ActiveDays:  []int{1, 2, 3, 4, 5},
			OpenHour:    9,
			OpenMinute:  0,
			CloseHour:   16,
			CloseMinute: 0,
			Aliases:     []string{"DFMGI", "FADGI", "FADX15"},
		},
		// Qatar Stock Exchange: Buffered 08:30 - 14:15 Doha time (Sunday - Thursday).
		{
			ID:          MarketQA,
			Timezone:    "Asia/Qatar",
			Location:    qatarLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    8,
			OpenMinute:  30,
			CloseHour:   14,
			CloseMinute: 15,
			Aliases:     []string{"QATAR", "QSE"},
		},
		// Qatar Indices: Aligned with Qatar main session.
		{
			ID:          MarketQAIndices,
			Timezone:    "Asia/Qatar",
			Location:    qatarLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    8,
			OpenMinute:  30,
			CloseHour:   14,
			CloseMinute: 15,
			Aliases:     []string{"QSE_INDEX"},
		},
		// Boursa Kuwait: Buffered 08:00 - 13:30 Kuwait time (Sunday - Thursday).
		{
			ID:          MarketKW,
			Timezone:    "Asia/Kuwait",
			Location:    kuwaitLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    8,
			OpenMinute:  0,
			CloseHour:   13,
			CloseMinute: 30,
			Aliases:     []string{"KUWAIT", "KSE"},
		},
		// Kuwait Indices: Aligned with Kuwait main session.
		{
			ID:          MarketKWIndices,
			Timezone:    "Asia/Kuwait",
			Location:    kuwaitLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    8,
			OpenMinute:  0,
			CloseHour:   13,
			CloseMinute: 30,
			Aliases:     []string{"BK_PREMIER", "BK_ALL_SHARE", "BK_MAIN"},
		},
		// Central Bank of Egypt Foreign Exchange Rates: 08:30 - 17:00 Cairo time (Sunday - Thursday).
		{
			ID:          MarketRates,
			Timezone:    "Africa/Cairo",
			Location:    cairoLoc,
			ActiveDays:  []int{0, 1, 2, 3, 4},
			OpenHour:    8,
			OpenMinute:  30,
			CloseHour:   17,
			CloseMinute: 0,
			Aliases:     []string{"BANKS", "BANK", "CBE", "CURRENCY", "CURRENCIES"},
		},
		// Egyptian Gold & Precious Metals: Monday - Saturday 09:00 - 22:00 Cairo time.
		{
			ID:          MarketGold,
			Timezone:    "Africa/Cairo",
			Location:    cairoLoc,
			ActiveDays:  []int{1, 2, 3, 4, 5, 6},
			OpenHour:    9,
			OpenMinute:  0,
			CloseHour:   22,
			CloseMinute: 0,
			Aliases:     []string{"METALS", "METAL"},
		},
		// Petroleum Fuel & Commodities: Aligned with gold schedule.
		{
			ID:          MarketCommodities,
			Timezone:    "Africa/Cairo",
			Location:    cairoLoc,
			ActiveDays:  []int{1, 2, 3, 4, 5, 6},
			OpenHour:    9,
			OpenMinute:  0,
			CloseHour:   22,
			CloseMinute: 0,
			Aliases:     []string{"COMMODITY", "GAS", "FUEL"},
		},
	}

	// Register configs, aliases, and holiday sets.
	for _, cfg := range rawConfigs {
		// Store configuration.
		engine.configs[cfg.ID] = cfg
		// Initialize empty holiday set.
		engine.holidays[cfg.ID] = make(map[string]bool)
		// Register ID in uppercase.
		engine.aliasLookup[strings.ToUpper(string(cfg.ID))] = cfg.ID
		// Register all configured aliases.
		for _, alias := range cfg.Aliases {
			// Register uppercase alias.
			engine.aliasLookup[strings.ToUpper(alias)] = cfg.ID
		}
	}

	// Return initialized engine pointer.
	return engine
}

// ResolveConfig looks up the MarketScheduleConfig for the requested market string or alias, defaulting to EGX.
func (e *MarketScheduleEngine) ResolveConfig(market string) *MarketScheduleConfig {
	// Normalize market input.
	norm := strings.ToUpper(strings.TrimSpace(market))
	// Check if string is empty.
	if norm == "" {
		// Default to EGX equities.
		return e.configs[MarketEGX]
	}

	// Look up canonical MarketID in alias table.
	if canonicalID, found := e.aliasLookup[norm]; found {
		// Return canonical config.
		if cfg, ok := e.configs[canonicalID]; ok {
			// Return matched config.
			return cfg
		}
	}

	// Fallback to EGX if unknown market.
	return e.configs[MarketEGX]
}

// RegisterHolidays registers holiday closure dates (YYYY-MM-DD) for a specific market exchange.
func (e *MarketScheduleEngine) RegisterHolidays(market string, holidays []string) {
	// Acquire write lock to serialize holiday map mutations.
	e.mu.Lock()
	// Defer release of write lock upon function return.
	defer e.mu.Unlock()

	// Resolve target market configuration.
	cfg := e.ResolveConfig(market)
	// Retrieve holiday set for this market.
	hSet, exists := e.holidays[cfg.ID]
	// If set is missing, initialize it.
	if !exists {
		// Allocate map.
		hSet = make(map[string]bool)
		// Store map.
		e.holidays[cfg.ID] = hSet
	}

	// Add each holiday date string to set.
	for _, h := range holidays {
		// Normalize date string.
		norm := strings.TrimSpace(h)
		// Insert holiday.
		if norm != "" {
			// Mark date as holiday.
			hSet[norm] = true
		}
	}
}

// isHoliday checks if the given YYYY-MM-DD date is a registered holiday for the market.
func (e *MarketScheduleEngine) isHoliday(id MarketID, dateStr string) bool {
	// Acquire read lock.
	e.mu.RLock()
	// Defer release of read lock upon function return.
	defer e.mu.RUnlock()

	// Look up market holiday set.
	if hSet, ok := e.holidays[id]; ok {
		// Return true if date exists in set.
		return hSet[dateStr]
	}

	// Return false by default.
	return false
}

// IsMarketOpen evaluates whether an exchange is actively trading at the specified reference time.
func (e *MarketScheduleEngine) IsMarketOpen(market string, refTime ...time.Time) bool {
	// Resolve target configuration.
	cfg := e.ResolveConfig(market)

	// Determine reference time (defaulting to time.Now() in UTC).
	t := time.Now().UTC()
	// Override if reference time provided.
	if len(refTime) > 0 && !refTime[0].IsZero() {
		// Assign reference time.
		t = refTime[0]
	}

	// Convert reference time to the target market's timezone.
	localTime := t.In(cfg.Location)
	// Format local ISO date string (YYYY-MM-DD).
	dateStr := localTime.Format("2006-01-02")
	// Extract integer weekday index (0=Sunday through 6=Saturday).
	dayOfWeek := int(localTime.Weekday())

	// Verify if today is an active trading weekday.
	isActiveDay := false
	// Check against configured active days.
	for _, d := range cfg.ActiveDays {
		// Compare weekday index.
		if d == dayOfWeek {
			// Found matching active day.
			isActiveDay = true
			// Break search loop.
			break
		}
	}

	// If today is a weekend or a holiday closure, market is closed.
	if !isActiveDay || e.isHoliday(cfg.ID, dateStr) {
		// Return false.
		return false
	}

	// Compute elapsed minutes since local midnight.
	currentMinutes := localTime.Hour()*60 + localTime.Minute()
	// Compute opening boundary in minutes.
	openMinutes := cfg.OpenHour*60 + cfg.OpenMinute
	// Compute closing boundary in minutes.
	closeMinutes := cfg.CloseHour*60 + cfg.CloseMinute

	// Check if current minutes fall inclusively within open and close window.
	return currentMinutes >= openMinutes && currentMinutes <= closeMinutes
}

// GetSecondsUntilNextOpen calculates the seconds remaining until the next opening bell (minimum 60s).
func (e *MarketScheduleEngine) GetSecondsUntilNextOpen(market string, refTime ...time.Time) int {
	// Resolve target configuration.
	cfg := e.ResolveConfig(market)

	// Determine reference time.
	t := time.Now().UTC()
	// Override if reference time provided.
	if len(refTime) > 0 && !refTime[0].IsZero() {
		// Assign reference time.
		t = refTime[0]
	}

	// Convert reference time to market timezone.
	localTime := t.In(cfg.Location)
	// Extract weekday.
	dayOfWeek := int(localTime.Weekday())
	// Extract date string.
	dateStr := localTime.Format("2006-01-02")
	// Compute elapsed minutes.
	currentMinutes := localTime.Hour()*60 + localTime.Minute()
	// Compute opening minutes.
	openMinutes := cfg.OpenHour*60 + cfg.OpenMinute

	// Check if market opens later today.
	isOpenDay := false
	// Verify active day.
	for _, d := range cfg.ActiveDays {
		// Match day.
		if d == dayOfWeek {
			// Active day confirmed.
			isOpenDay = true
			// Break loop.
			break
		}
	}

	// Check if market can open today.
	canOpenToday := isOpenDay && !e.isHoliday(cfg.ID, dateStr) && currentMinutes < openMinutes

	// Target date for next open.
	targetDate := localTime
	// If market does not open today, step forward day by day.
	if !canOpenToday {
		// Step forward up to 14 days.
		for step := 1; step <= 14; step++ {
			// Advance candidate date by 24 hours.
			candidate := localTime.Add(time.Duration(step*24) * time.Hour)
			// Extract candidate weekday.
			cDay := int(candidate.Weekday())
			// Extract candidate date string.
			cDateStr := candidate.Format("2006-01-02")

			// Check if candidate is active trading day.
			cActive := false
			// Iterate active days.
			for _, d := range cfg.ActiveDays {
				// Match day.
				if d == cDay {
					// Active day confirmed.
					cActive = true
					// Break loop.
					break
				}
			}

			// Check if active day and not a holiday.
			if cActive && !e.isHoliday(cfg.ID, cDateStr) {
				// Select candidate date.
				targetDate = candidate
				// Break search loop.
				break
			}
		}
	}

	// Construct opening time on target date in market timezone.
	nextOpenTime := time.Date(
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		cfg.OpenHour,
		cfg.OpenMinute,
		0,
		0,
		cfg.Location,
	)

	// Calculate difference in seconds.
	diffSeconds := int(nextOpenTime.Sub(localTime).Seconds())
	// Apply defensive floor of 60 seconds.
	if diffSeconds < 60 {
		// Floor at 60s.
		diffSeconds = 60
	}

	// Return remaining seconds.
	return diffSeconds
}

// GetLatestTradingSessionDate deterministically returns the YYYY-MM-DD date of the most recent active or completed session.
func (e *MarketScheduleEngine) GetLatestTradingSessionDate(market string, refTime ...time.Time) string {
	// Resolve target configuration.
	cfg := e.ResolveConfig(market)

	// Determine reference time.
	t := time.Now().UTC()
	// Override if provided.
	if len(refTime) > 0 && !refTime[0].IsZero() {
		// Assign reference time.
		t = refTime[0]
	}

	// Convert to market timezone.
	localTime := t.In(cfg.Location)
	// Extract weekday.
	dayOfWeek := int(localTime.Weekday())
	// Extract date string.
	dateStr := localTime.Format("2006-01-02")
	// Compute current minutes.
	currentMinutes := localTime.Hour()*60 + localTime.Minute()
	// Compute open minutes.
	openMinutes := cfg.OpenHour*60 + cfg.OpenMinute

	// Check if today is active day.
	isActiveDay := false
	// Compare days.
	for _, d := range cfg.ActiveDays {
		// Match day.
		if d == dayOfWeek {
			// Active day.
			isActiveDay = true
			// Break loop.
			break
		}
	}

	// If today is a trading day, not a holiday, and opening time has arrived.
	if isActiveDay && !e.isHoliday(cfg.ID, dateStr) && currentMinutes >= openMinutes {
		// Today is the active or latest session date.
		return dateStr
	}

	// Otherwise step backward in time up to 14 days.
	for step := 1; step <= 14; step++ {
		// Decrement candidate by step days.
		candidate := localTime.Add(-time.Duration(step*24) * time.Hour)
		// Extract candidate weekday.
		cDay := int(candidate.Weekday())
		// Extract candidate date string.
		cDateStr := candidate.Format("2006-01-02")

		// Check active day.
		cActive := false
		// Compare days.
		for _, d := range cfg.ActiveDays {
			// Match day.
			if d == cDay {
				// Active day.
				cActive = true
				// Break loop.
				break
			}
		}

		// If candidate is active day and not holiday.
		if cActive && !e.isHoliday(cfg.ID, cDateStr) {
			// Return resolved trading date.
			return cDateStr
		}
	}

	// Defensive fallback to reference date.
	return dateStr
}

// GetMarketState returns the full operational evaluation envelope for the requested market.
func (e *MarketScheduleEngine) GetMarketState(market string, refTime ...time.Time) MarketState {
	// Resolve target configuration.
	cfg := e.ResolveConfig(market)

	// Determine reference time.
	t := time.Now().UTC()
	// Override if provided.
	if len(refTime) > 0 && !refTime[0].IsZero() {
		// Assign reference time.
		t = refTime[0]
	}

	// Convert to market timezone.
	localTime := t.In(cfg.Location)
	// Calculate decimal hour.
	decimalHour := float64(localTime.Hour()) + float64(localTime.Minute())/60.0 + float64(localTime.Second())/3600.0
	// Check open status.
	isOpen := e.IsMarketOpen(market, t)
	// Calculate seconds until next open.
	secondsUntilOpen := e.GetSecondsUntilNextOpen(market, t)
	// Compute next open epoch milliseconds.
	nextOpenMs := t.UnixMilli() + int64(secondsUntilOpen*1000)
	// Resolve latest trading session date.
	latestSessionDate := e.GetLatestTradingSessionDate(market, t)

	// Return populated MarketState structure.
	return MarketState{
		// Canonical market identifier.
		Market: cfg.ID,
		// Local date string.
		CurrentLocalDate: localTime.Format("2006-01-02"),
		// Local day of week index.
		LocalDayOfWeek: int(localTime.Weekday()),
		// Local hour decimal.
		LocalHourDecimal: decimalHour,
		// Boolean open status.
		IsOpen: isOpen,
		// Next open timestamp milliseconds.
		NextOpenTimestampMs: nextOpenMs,
		// Seconds until next open.
		SecondsUntilNextOpen: secondsUntilOpen,
		// Latest trading session date.
		LatestTradingSessionDate: latestSessionDate,
	}
}

// ComputeCacheTTL returns dynamic Cache-Control TTL in seconds: 60s during trading; seconds until next open when closed.
func (e *MarketScheduleEngine) ComputeCacheTTL(market string, refTime ...time.Time) int {
	// If market is currently open.
	if e.IsMarketOpen(market, refTime...) {
		// Enforce 60s fast caching during active trading session.
		return 60
	}

	// If closed, return duration until next opening bell (capped at 86400s / 24 hours).
	ttl := e.GetSecondsUntilNextOpen(market, refTime...)
	// Cap maximum TTL at 24 hours to ensure periodic client refresh.
	if ttl > 86400 {
		// Cap at 86400.
		ttl = 86400
	}

	// Return computed off-hours TTL.
	return ttl
}

// NowCairo returns the current operational wall-clock time in Africa/Cairo timezone.
func (e *MarketScheduleEngine) NowCairo() time.Time {
	// Return current time converted to Africa/Cairo location.
	return time.Now().In(e.cairoLoc)
}

// CairoLocation returns the cached *time.Location pointer for Africa/Cairo.
func (e *MarketScheduleEngine) CairoLocation() *time.Location {
	// Return cached Africa/Cairo location.
	return e.cairoLoc
}

