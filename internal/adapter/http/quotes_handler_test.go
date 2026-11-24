// Package http_test provides integration tests for inbound HTTP handlers and routing.
package http_test

// Import standard library packages for testing and JSON decoding.
import (
	// bytes provides request body buffer.
	"bytes"
	// encoding/json provides JSON decoding.
	"encoding/json"
	// net/http provides request primitives.
	"net/http"
	// net/http/httptest provides response recorder.
	"net/http/httptest"
	// testing provides test fixtures.
	"testing"
	// time provides temporal calculations.
	"time"

	// adapterhttp provides HTTP handlers.
	adapterhttp "bkam-api/internal/adapter/http"
	// domain defines quote types.
	"bkam-api/internal/domain"
	// engine provides in-memory coordinators.
	"bkam-api/internal/engine"
)

// setupQuotesTestHandler configures a test QuotesHandler with pre-seeded in-memory quotes.
func setupQuotesTestHandler() *adapterhttp.QuotesHandler {
	// Construct coordinator.
	coord := engine.NewLiveSnapshotCoordinator(nil)

	// Pre-seed bank currency rates.
	coord.UpdateSnapshot("rates", &domain.RateSnapshot{
		RequestedAt: time.Now().UTC(),
		Rates: []domain.CurrencyRate{
			{
				BankCode:     "CBE",
				CurrencyCode: "USD",
				NameEn:       "Central Bank of Egypt",
				BuyRate:      48.50,
				SellRate:     48.65,
				RetrievedAt:  time.Now().UTC(),
			},
		},
	})

	// Old sell price pointer.
	oldSell := 3510.0
	// Pre-seed gold prices.
	coord.UpdateSnapshot("gold", &domain.GoldSnapshot{
		UpdatedAt: time.Now().UTC(),
		Currency:  "EGP",
		Rates: map[string]domain.GoldKaratPrice{
			"21k": {Buy: 3500.0, Sell: 3520.0, OldSell: &oldSell},
		},
	})

	// Pre-seed EGX stocks.
	coord.UpdateSnapshot("stocks", &domain.StockSnapshot{
		Market:    "egx",
		UpdatedAt: time.Now().UTC(),
		Currency:  "EGP",
		Stocks: map[string]domain.StockPriceItem{
			"COMI": {Symbol: "COMI", NameEn: "Commercial International Bank", Price: 85.0},
		},
	}, "egx")

	// Pre-seed EGX indices.
	coord.UpdateSnapshot("indices", &domain.IndexSnapshot{
		Market:    "egx",
		UpdatedAt: time.Now().UTC(),
		Currency:  "EGP",
		Indices: map[string]domain.IndexPriceItem{
			"EGX30": {IndexCode: "EGX30", Name: "EGX 30 Index", Price: 30500.0},
		},
	}, "egx")

	// Construct and return handler.
	return adapterhttp.NewQuotesHandler(coord)
}

// TestBatchQuotesResolution verifies resolution of heterogeneous asset identifiers and unresolved tracking.
func TestBatchQuotesResolution(t *testing.T) {
	// Setup handler.
	h := setupQuotesTestHandler()

	// Prepare POST request body with mixed identifiers and one invalid asset.
	bodyBytes, _ := json.Marshal(map[string]any{
		"items": []string{"CBE_USD", "21k", "COMI", "EGX30", "NONEXISTENT_ITEM"},
	})
	// Create POST request.
	req, _ := http.NewRequest("POST", "/v1/quotes", bytes.NewReader(bodyBytes))
	// Recorder.
	rec := httptest.NewRecorder()

	// Execute handler.
	h.PostQuotes(rec, req)

	// Assert HTTP 200 OK.
	if rec.Code != http.StatusOK {
		// Report error.
		t.Fatalf("expected HTTP 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Unmarshal response.
	var res domain.BatchQuoteResponse
	// Decode JSON.
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		// Report decode failure.
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify resolved quotes count.
	if len(res.Quotes) != 4 {
		// Report error.
		t.Errorf("expected 4 resolved quotes, got %d", len(res.Quotes))
	}

	// Verify unresolved items count.
	if len(res.Unresolved) != 1 || res.Unresolved[0] != "NONEXISTENT_ITEM" {
		// Report error.
		t.Errorf("expected unresolved [NONEXISTENT_ITEM], got %v", res.Unresolved)
	}

	// Verify CBE_USD quote.
	if usdQuote, found := res.Quotes["CBE_USD"]; !found || usdQuote.Price != 48.65 {
		// Report error.
		t.Errorf("expected CBE_USD price 48.65, got %v", usdQuote)
	}

	// Verify 21k gold quote.
	if goldQuote, found := res.Quotes["21k"]; !found || goldQuote.Price != 3520.0 {
		// Report error.
		t.Errorf("expected 21k price 3520.0, got %v", goldQuote)
	}
}

// TestBatchQuotesGetQueryParams verifies resolution using GET query parameters.
func TestBatchQuotesGetQueryParams(t *testing.T) {
	// Setup handler.
	h := setupQuotesTestHandler()

	// Create GET request.
	req, _ := http.NewRequest("GET", "/v1/quotes?items=CBE_USD,21k", nil)
	// Recorder.
	rec := httptest.NewRecorder()

	// Execute handler.
	h.GetQuotes(rec, req)

	// Assert HTTP 200 OK.
	if rec.Code != http.StatusOK {
		// Report error.
		t.Fatalf("expected HTTP 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Unmarshal response.
	var res domain.BatchQuoteResponse
	// Decode JSON.
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		// Report failure.
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify resolved count.
	if len(res.Quotes) != 2 {
		// Report error.
		t.Errorf("expected 2 resolved quotes, got %d", len(res.Quotes))
	}
}

// TestBatchQuotesValidationErrors verifies rejection of empty requests.
func TestBatchQuotesValidationErrors(t *testing.T) {
	// Setup handler.
	h := setupQuotesTestHandler()

	// Case 1: Empty POST body items.
	bodyBytes, _ := json.Marshal(map[string]any{"items": []string{}})
	// Request.
	req1, _ := http.NewRequest("POST", "/v1/quotes", bytes.NewReader(bodyBytes))
	// Recorder.
	rec1 := httptest.NewRecorder()
	// Execute.
	h.PostQuotes(rec1, req1)
	// Assert 400.
	if rec1.Code != http.StatusBadRequest {
		// Report error.
		t.Errorf("expected HTTP 400 for empty POST items, got %d", rec1.Code)
	}

	// Case 2: Missing GET query items.
	req2, _ := http.NewRequest("GET", "/v1/quotes", nil)
	// Recorder.
	rec2 := httptest.NewRecorder()
	// Execute.
	h.GetQuotes(rec2, req2)
	// Assert 400.
	if rec2.Code != http.StatusBadRequest {
		// Report error.
		t.Errorf("expected HTTP 400 for missing GET items, got %d", rec2.Code)
	}
}
