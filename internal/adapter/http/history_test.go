// Package http_test provides integration tests for inbound HTTP handlers and routing.
package http_test

// Import standard library packages for HTTP testing.
import (
	// net/http provides request primitives.
	"net/http"
	// net/http/httptest provides response recorder.
	"net/http/httptest"
	// testing provides Go testing fixtures.
	"testing"
	// time provides temporal calculations.
	"time"

	// adapterhttp provides HTTP router and handlers.
	adapterhttp "bkam-api/internal/adapter/http"
	// engine provides coordinators and engines.
	"bkam-api/internal/engine"
)

// setupHistoryTestRouter configures an in-memory test environment including HistoryResolutionEngine.
func setupHistoryTestRouter() http.Handler {
	// Construct coordinator.
	coord := engine.NewLiveSnapshotCoordinator(nil)
	// Construct schedule engine.
	sched := engine.NewMarketScheduleEngine()
	// Construct history engine.
	histEngine := engine.NewHistoryResolutionEngine(nil, coord, sched)

	// Pre-seed an active weekly bucket in RAM.
	histEngine.SetActiveWeeklyBucket("rates_history_weekly", "2026-W40", map[string]any{
		"2026-10-01": map[string]any{"rates": []any{}},
		"2026-10-02": map[string]any{"rates": []any{}},
	})
	histEngine.SetActiveWeeklyBucket("gold_prices_history_weekly", "2026-W40", map[string]any{
		"2026-10-01": map[string]any{"rates": map[string]any{}},
	})
	histEngine.SetActiveWeeklyBucket("stock_prices_history_weekly", "2026-W40", map[string]any{
		"2026-10-01": map[string]any{"stocks": map[string]any{}},
	})
	histEngine.SetActiveWeeklyBucket("index_prices_history_weekly", "2026-W40", map[string]any{
		"2026-10-01": map[string]any{"indices": map[string]any{}},
	})

	// Construct handlers with history engine.
	ratesH := adapterhttp.NewRatesHandler(coord, histEngine)
	commoditiesH := adapterhttp.NewCommoditiesHandler(coord, histEngine)
	stockH := adapterhttp.NewStockHandler(coord, histEngine)
	indicesH := adapterhttp.NewIndicesHandler(coord, histEngine)

	// Return configured router.
	return adapterhttp.NewRouter(&adapterhttp.RouterDependencies{
		Coordinator:        coord,
		MarketSchedule:     sched,
		RatesHandler:       ratesH,
		CommoditiesHandler: commoditiesH,
		StockHandler:       stockH,
		IndicesHandler:     indicesH,
	})
}

// TestHistoricalEndpoints verifies resolution of historical single-date and range queries.
func TestHistoricalEndpoints(t *testing.T) {
	// Setup test router.
	router := setupHistoryTestRouter()

	// List of historical test endpoints to probe.
	testCases := []struct {
		url          string
		expectedCode int
	}{
		{"/v1/rates/history/2026-10-01", http.StatusOK},
		{"/v1/rates/history/range?start=2026-10-01&end=2026-10-02", http.StatusOK},
		{"/v1/gold/history/2026-10-01", http.StatusOK},
		{"/v1/gold/history/range?start=2026-10-01&end=2026-10-01", http.StatusOK},
		{"/v1/stocks/history/2026-10-01", http.StatusOK},
		{"/v1/stocks/COMI/history/range?start=2026-10-01&end=2026-10-01", http.StatusOK},
		{"/v1/indices/history/2026-10-01", http.StatusOK},
		{"/v1/indices/EGX30/history/range?start=2026-10-01&end=2026-10-01", http.StatusOK},
	}

	// Iterate test cases.
	for _, tc := range testCases {
		// Create request.
		req, _ := http.NewRequest("GET", tc.url, nil)
		// Recorder.
		rec := httptest.NewRecorder()
		// Execute request.
		router.ServeHTTP(rec, req)

		// Assert status.
		if rec.Code != tc.expectedCode {
			// Error.
			t.Errorf("expected route %s to return %d, got %d: %s", tc.url, tc.expectedCode, rec.Code, rec.Body.String())
		}
	}
}

// TestHistoricalFutureDateRejection verifies that future date queries are rejected with HTTP 400.
func TestHistoricalFutureDateRejection(t *testing.T) {
	// Setup router.
	router := setupHistoryTestRouter()

	// Compute future date.
	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	// Target URL.
	url := "/v1/rates/history/" + futureDate

	// Create request.
	req, _ := http.NewRequest("GET", url, nil)
	// Recorder.
	rec := httptest.NewRecorder()
	// Execute.
	router.ServeHTTP(rec, req)

	// Assert 400 Bad Request.
	if rec.Code != http.StatusBadRequest {
		// Report error.
		t.Errorf("expected future date query %s to return 400, got %d", url, rec.Code)
	}
}
