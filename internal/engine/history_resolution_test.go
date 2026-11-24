// Package engine_test provides unit tests for the HistoryResolutionEngine.
package engine_test

// Import standard library packages for temporary filesystem, testing, and time.
import (
	// os provides temporary directory cleanup.
	"os"
	// testing provides Go testing primitives.
	"testing"
	// time provides temporal calculations.
	"time"

	// storage provides the local filesystem persistence adapter.
	"bkam-api/internal/adapter/storage"
	// domain defines core financial models.
	"bkam-api/internal/domain"
	// engine provides HistoryResolutionEngine and coordinators.
	"bkam-api/internal/engine"
)

// TestHistoryResolutionEngineTier1AndTier2 verifies multi-tier lookup across RAM and Disk.
func TestHistoryResolutionEngineTier1AndTier2(t *testing.T) {
	// Create temporary directory for disk storage.
	tmpDir, err := os.MkdirTemp("", "history-engine-test-*")
	// Verify directory creation.
	if err != nil {
		// Log error.
		t.Fatalf("failed to create temp dir: %v", err)
	}
	// Defer cleanup.
	defer os.RemoveAll(tmpDir)

	// Construct disk adapter.
	disk, _ := storage.NewDiskAdapter(tmpDir)
	// Construct coordinator.
	coord := engine.NewLiveSnapshotCoordinator(disk)
	// Construct schedule engine.
	sched := engine.NewMarketScheduleEngine()
	// Construct history resolution engine.
	histEngine := engine.NewHistoryResolutionEngine(disk, coord, sched)

	// Sample historical weekly bucket for 2026-W40.
	pastWeekData := map[string]any{
		"2026-10-01": map[string]any{"price": 3500.0, "status": "sealed"},
		"2026-10-02": map[string]any{"price": 3510.0, "status": "sealed"},
	}

	// Persist sealed past week to Tier 2 disk storage.
	if err := disk.SaveHistoryBucket("gold_prices_history_weekly", "2026-W40", pastWeekData); err != nil {
		// Log error.
		t.Fatalf("failed to save historical bucket to disk: %v", err)
	}

	// Query historical date snapshot for 2026-10-01.
	resolved, err := histEngine.GetDateSnapshot("gold", "2026-10-01", domain.RangeOptions{})
	// Verify no error.
	if err != nil {
		// Log error.
		t.Fatalf("unexpected error fetching historical date: %v", err)
	}

	// Verify resolved snapshot.
	if resolved == nil || resolved.Data == nil {
		// Fail assertion.
		t.Fatalf("expected resolved snapshot for 2026-10-01")
	}

	// Assert not today.
	if resolved.IsToday {
		// Error.
		t.Errorf("expected IsToday to be false for sealed past date")
	}

	// Assert CacheTTLSeconds is 3600 (1 hour for historical archives).
	if resolved.CacheTTLSeconds != 3600 {
		// Error.
		t.Errorf("expected CacheTTLSeconds to be 3600, got %d", resolved.CacheTTLSeconds)
	}
}

// TestHistoryResolutionEngineRangeSlicing verifies multi-day range assembly across weekly buckets.
func TestHistoryResolutionEngineRangeSlicing(t *testing.T) {
	// Create temporary directory.
	tmpDir, _ := os.MkdirTemp("", "history-range-*")
	// Cleanup on return.
	defer os.RemoveAll(tmpDir)

	// Construct dependencies.
	disk, _ := storage.NewDiskAdapter(tmpDir)
	// Construct coordinator.
	coord := engine.NewLiveSnapshotCoordinator(disk)
	// Construct schedule.
	sched := engine.NewMarketScheduleEngine()
	// Construct history engine.
	histEngine := engine.NewHistoryResolutionEngine(disk, coord, sched)

	// Populate Tier 1 active RAM bucket for 2026-W45.
	weekBucket := map[string]any{
		"2026-11-03": map[string]any{"rate": 49.20},
		"2026-11-04": map[string]any{"rate": 49.30},
		"2026-11-05": map[string]any{"rate": 49.40},
	}
	// Register active week in RAM.
	histEngine.SetActiveWeeklyBucket("rates_history_weekly", "2026-W45", weekBucket)

	// Query 3-day range.
	results, err := histEngine.GetDateRange("rates", domain.RangeOptions{
		StartDate: "2026-11-03",
		EndDate:   "2026-11-05",
	})
	// Assert no error.
	if err != nil {
		// Log error.
		t.Fatalf("failed to get date range: %v", err)
	}

	// Assert 3 items resolved.
	if len(results) != 3 {
		// Report error.
		t.Errorf("expected 3 items in range, got %d", len(results))
	}
}

// TestHistoryResolutionEngineFutureDateGuard verifies that future dates are rejected with validation error.
func TestHistoryResolutionEngineFutureDateGuard(t *testing.T) {
	// Construct history engine.
	histEngine := engine.NewHistoryResolutionEngine(nil, nil, engine.NewMarketScheduleEngine())

	// Compute far future date.
	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	// Attempt querying future date.
	_, err := histEngine.GetDateSnapshot("rates", futureDate, domain.RangeOptions{})
	// Assert error is returned.
	if err == nil {
		// Report error.
		t.Errorf("expected future date query to fail with validation error")
	}
}

// TestHistoryResolutionEngineSaudiNumericCodeRange verifies that regional Saudi stock ranges match by numeric code.
func TestHistoryResolutionEngineSaudiNumericCodeRange(t *testing.T) {
	// Construct history engine with in-memory dependencies.
	histEngine := engine.NewHistoryResolutionEngine(nil, nil, engine.NewMarketScheduleEngine())

	// Mock weekly bucket for 2026-W45 containing Sunday 2026-11-08.
	w45Bucket := map[string]any{
		// Sunday trading day.
		"2026-11-08": map[string]any{
			// Stocks dictionary.
			"stocks": map[string]any{
				// Al Rajhi Bank keyed by alphanumeric ticker.
				"ALRAJHI": map[string]any{
					// Standard ticker symbol.
					"symbol": "ALRAJHI",
					// Numeric exchange code.
					"numericCode": "1120",
					// Trading price.
					"price": 63.95,
					// Price change.
					"change": -0.40,
					// Change percentage.
					"changePercent": -0.62,
					// Trading currency.
					"currency": "SAR",
				},
			},
		},
	}

	// Mock weekly bucket for 2026-W46 containing Monday 2026-11-09 and Tuesday 2026-11-10.
	w46Bucket := map[string]any{
		// Monday trading day.
		"2026-11-09": map[string]any{
			// Stocks dictionary.
			"stocks": map[string]any{
				// Al Rajhi Bank.
				"ALRAJHI": map[string]any{
					// Standard ticker symbol.
					"symbol": "ALRAJHI",
					// Numeric exchange code.
					"numericCode": "1120",
					// Trading price.
					"price": 64.20,
					// Price change.
					"change": 0.25,
					// Change percentage.
					"changePercent": 0.39,
					// Trading currency.
					"currency": "SAR",
				},
			},
		},
		// Tuesday trading day.
		"2026-11-10": map[string]any{
			// Stocks dictionary.
			"stocks": map[string]any{
				// Al Rajhi Bank.
				"ALRAJHI": map[string]any{
					// Standard ticker symbol.
					"symbol": "ALRAJHI",
					// Numeric exchange code.
					"numericCode": "1120",
					// Trading price.
					"price": 64.50,
					// Price change.
					"change": 0.30,
					// Change percentage.
					"changePercent": 0.47,
					// Trading currency.
					"currency": "SAR",
				},
			},
		},
	}

	// Register 2026-W45 active weekly bucket in memory.
	histEngine.SetActiveWeeklyBucket("stock_prices_saudi_history_weekly", "2026-W45", w45Bucket)
	// Register 2026-W46 active weekly bucket in memory.
	histEngine.SetActiveWeeklyBucket("stock_prices_saudi_history_weekly", "2026-W46", w46Bucket)

	// Query historical range spanning across W45 and W46 using numeric ticker 1120.
	results, err := histEngine.GetDateRange("stocks", domain.RangeOptions{
		// Start date in W45.
		StartDate: "2026-11-08",
		// End date in W46.
		EndDate: "2026-11-10",
		// Saudi market.
		Market: "saudi",
		// Target numeric ticker code.
		Symbol: "1120",
	})
	// Assert no error returned.
	if err != nil {
		// Fatal test failure.
		t.Fatalf("failed to resolve range for numeric ticker 1120: %v", err)
	}

	// Assert exactly 3 trading day records were matched across the weekly boundary.
	if len(results) != 3 {
		// Report mismatch.
		t.Fatalf("expected 3 historical records for ticker 1120, got %d", len(results))
	}

	// Inspect first record.
	firstRecord, ok := results[0].(map[string]any)
	// Assert map cast succeeds.
	if !ok {
		// Report type error.
		t.Fatalf("expected map[string]any record, got %T", results[0])
	}

	// Assert price matches.
	if price, _ := firstRecord["price"].(float64); price != 63.95 {
		// Report mismatch.
		t.Errorf("expected price 63.95, got %v", firstRecord["price"])
	}
}
