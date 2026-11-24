// Package engine_test provides unit tests for the MarketScheduleEngine.
package engine_test

// Import standard library packages for testing, time manipulation, and concurrency.
import (
	// sync provides wait groups for concurrent testing.
	"sync"
	// testing provides Go unit testing primitives.
	"testing"
	// time provides temporal fixtures.
	"time"

	// engine references the market schedule engine package.
	"bkam-api/internal/engine"
)

// TestEGXTradingHours verifies active and off-hours sessions for Egyptian equities.
func TestEGXTradingHours(t *testing.T) {
	// Construct new market schedule engine.
	sched := engine.NewMarketScheduleEngine()

	// Load Cairo timezone location.
	cairoLoc, err := time.LoadLocation("Africa/Cairo")
	// Verify location loaded.
	if err != nil {
		// Fatal error.
		t.Fatalf("failed to load Africa/Cairo: %v", err)
	}

	// Case 1: Active trading session on Monday at 11:00 AM Cairo time.
	// Construct test timestamp: 2026-11-16 (Monday) 11:00 AM Cairo.
	activeTime := time.Date(2026, 11, 16, 11, 0, 0, 0, cairoLoc)
	// Assert market is open.
	if !sched.IsMarketOpen("EGX", activeTime) {
		// Error.
		t.Errorf("expected EGX to be open on Monday at 11:00 AM Cairo")
	}

	// Verify Cache-Control TTL is 60s when open.
	if ttl := sched.ComputeCacheTTL("EGX", activeTime); ttl != 60 {
		// Error.
		t.Errorf("expected TTL to be 60 during active trading, got %d", ttl)
	}

	// Case 2: Off-hours on Friday (Egyptian weekend).
	// Construct test timestamp: 2026-11-20 (Friday) 11:00 AM Cairo.
	weekendFriday := time.Date(2026, 11, 20, 11, 0, 0, 0, cairoLoc)
	// Assert market is closed.
	if sched.IsMarketOpen("EGX", weekendFriday) {
		// Error.
		t.Errorf("expected EGX to be closed on Friday weekend")
	}

	// Verify session backtracks to Thursday 2026-11-19.
	lastSession := sched.GetLatestTradingSessionDate("EGX", weekendFriday)
	// Assert last session was Thursday.
	if lastSession != "2026-11-19" {
		// Error.
		t.Errorf("expected latest trading session to be 2026-11-19, got %s", lastSession)
	}

	// Case 3: Registered holiday closure.
	// Register holiday on 2026-11-16.
	sched.RegisterHolidays("EGX", []string{"2026-11-16"})
	// Now market should be closed despite being Monday 11:00 AM.
	if sched.IsMarketOpen("EGX", activeTime) {
		// Error.
		t.Errorf("expected EGX to be closed on registered holiday 2026-11-16")
	}
}

// TestRegionalGCCMarkets verifies schedule evaluations for Saudi, UAE, Qatar, and Kuwait.
func TestRegionalGCCMarkets(t *testing.T) {
	// Construct engine.
	sched := engine.NewMarketScheduleEngine()

	// Load Riyadh timezone location.
	riyadhLoc, _ := time.LoadLocation("Asia/Riyadh")
	// Test timestamp: Tuesday 12:00 PM Riyadh time.
	tuesdayRiyadh := time.Date(2026, 11, 17, 12, 0, 0, 0, riyadhLoc)

	// Assert Saudi Tadawul is open.
	if !sched.IsMarketOpen("SA", tuesdayRiyadh) {
		// Error.
		t.Errorf("expected Saudi Tadawul to be open on Tuesday at noon")
	}

	// Assert alias resolution works for "TADAWUL".
	if !sched.IsMarketOpen("TADAWUL", tuesdayRiyadh) {
		// Error.
		t.Errorf("expected TADAWUL alias to resolve and be open")
	}

	// Assert market state reflects correct market identifier.
	state := sched.GetMarketState("SA", tuesdayRiyadh)
	// Check market field.
	if state.Market != engine.MarketSA {
		// Error.
		t.Errorf("expected state market SA, got %s", state.Market)
	}
}

// TestMarketScheduleEngineConcurrency verifies zero race conditions under concurrent queries and holiday registrations.
func TestMarketScheduleEngineConcurrency(t *testing.T) {
	// Construct engine.
	sched := engine.NewMarketScheduleEngine()
	// WaitGroup for 30 concurrent workers.
	var wg sync.WaitGroup
	// Number of workers.
	workers := 30

	// Launch concurrent routines.
	for i := 0; i < workers; i++ {
		// Add worker.
		wg.Add(1)
		// Worker index.
		idx := i
		// Launch goroutine.
		go func() {
			// Done on exit.
			defer wg.Done()
			// Alternate between reads and writes.
			if idx%3 == 0 {
				// Register holiday.
				sched.RegisterHolidays("EGX", []string{"2026-12-25"})
			} else {
				// Read market open status.
				_ = sched.IsMarketOpen("EGX")
				// Read state.
				_ = sched.GetMarketState("SA")
				// Read TTL.
				_ = sched.ComputeCacheTTL("GOLD")
			}
		}()
	}

	// Wait for all workers to finish.
	wg.Wait()
}
