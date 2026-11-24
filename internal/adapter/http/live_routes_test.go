// Package http_test provides integration tests for inbound HTTP handlers and routing.
package http_test

// Import standard library packages for HTTP recording and testing.
import (
	// net/http provides request primitives.
	"net/http"
	// net/http/httptest provides response recorder for offline HTTP testing.
	"net/http/httptest"
	// testing provides Go testing primitives.
	"testing"

	// adapter/http provides handlers and router.
	adapterhttp "bkam-api/internal/adapter/http"
	// engine provides coordinators.
	"bkam-api/internal/engine"
)

// setupTestRouter instantiates an in-memory coordinator and configured router for testing.
func setupTestRouter() http.Handler {
	// Construct coordinator without disk adapter.
	coord := engine.NewLiveSnapshotCoordinator(nil)
	// Construct market schedule engine.
	sched := engine.NewMarketScheduleEngine()
	// Construct intraday timeline store.
	intradayStore := engine.NewIntradayTimelineStore(sched, coord, nil)

	// Construct handlers.
	ratesH := adapterhttp.NewRatesHandler(coord, nil)
	ratesH.SetIntradayStore(intradayStore)
	commoditiesH := adapterhttp.NewCommoditiesHandler(coord, nil)
	commoditiesH.SetIntradayStore(intradayStore)
	stockH := adapterhttp.NewStockHandler(coord, nil)
	stockH.SetIntradayStore(intradayStore)
	indicesH := adapterhttp.NewIndicesHandler(coord, nil)
	quotesH := adapterhttp.NewQuotesHandler(coord)

	// Assemble router.
	return adapterhttp.NewRouter(&adapterhttp.RouterDependencies{
		Coordinator:        coord,
		MarketSchedule:     sched,
		QuotesHandler:      quotesH,
		RatesHandler:       ratesH,
		CommoditiesHandler: commoditiesH,
		StockHandler:       stockH,
		IndicesHandler:     indicesH,
	})
}

// TestLiveQuotesEndpoints verifies that all live quote routes respond with HTTP 200 OK.
func TestLiveQuotesEndpoints(t *testing.T) {
	// Setup test router.
	router := setupTestRouter()

	// List of test endpoints to probe.
	routes := []string{
		"/v1/catalog",
		"/v1/rates",
		"/v1/rates/USD",
		"/v1/rates/USD/intraday/latest",
		"/v1/gold",
		"/v1/gold/intraday/latest",
		"/v1/gold/21/intraday/latest",
		"/v1/gas",
		"/v1/stocks",
		"/v1/stocks/COMI/intraday/latest",
		"/v1/stocks/saudi",
		"/v1/stocks/regional/catalog",
		"/v1/indices",
	}

	// Iterate through each test endpoint.
	for _, route := range routes {
		// Create mock GET request.
		req, err := http.NewRequest("GET", route, nil)
		// Verify request created.
		if err != nil {
			// Fatal error.
			t.Fatalf("failed to create request for %s: %v", route, err)
		}

		// Create recorder.
		rec := httptest.NewRecorder()
		// Serve HTTP request.
		router.ServeHTTP(rec, req)

		// Assert status code is 200 OK.
		if rec.Code != http.StatusOK {
			// Report error.
			t.Errorf("expected route %s to return 200 OK, got %d: %s", route, rec.Code, rec.Body.String())
		}
	}
}

// TestRatesInvalidCurrencyFilter verifies that unsupported currencies return 400 INVALID_FILTER.
func TestRatesInvalidCurrencyFilter(t *testing.T) {
	// Setup router.
	router := setupTestRouter()

	// Create request with invalid currency XYZ.
	req, _ := http.NewRequest("GET", "/v1/rates/XYZ", nil)
	// Recorder.
	rec := httptest.NewRecorder()
	// Serve request.
	router.ServeHTTP(rec, req)

	// Assert 400 Bad Request.
	if rec.Code != http.StatusBadRequest {
		// Report error.
		t.Errorf("expected 400 Bad Request for unsupported currency, got %d", rec.Code)
	}
}
