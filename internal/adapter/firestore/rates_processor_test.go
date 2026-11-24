// Package firestore_test provides unit tests for the Firestore rates processor.
package firestore_test

// Import standard library packages for testing.
import (
	// testing provides Go unit test fixtures.
	"testing"

	// firestore points to the processor package under test.
	"bkam-api/internal/adapter/firestore"
	// domain defines core rate models.
	"bkam-api/internal/domain"
)

// TestRatesProcessorEnrichment verifies that catalog metadata (Arabic/English names, logos) are enriched into rates.
func TestRatesProcessorEnrichment(t *testing.T) {
	// Construct new rates processor.
	processor := firestore.NewRatesProcessor()

	// Set sample old rate for trend calculation.
	oldBuy := 49.60
	// Construct raw bank document input matching Firestore structure.
	rawDocs := []firestore.RawBankDocument{
		{
			BankID: "nbe",
			Status: "SUCCESS",
			Rates: map[string]firestore.RawBankRateInfo{
				"USD": {
					Buy:    49.50,
					Sell:   49.60,
					OldBuy: &oldBuy,
				},
			},
		},
	}

	// Process raw documents.
	snapshot := processor.ProcessRawDocuments(rawDocs, "test-trace-123")

	// Assert snapshot not nil.
	if snapshot == nil {
		// Fatal error.
		t.Fatalf("expected snapshot to be returned")
	}

	// Assert rate count is 1.
	if len(snapshot.Rates) != 1 {
		// Error.
		t.Fatalf("expected 1 rate, got %d", len(snapshot.Rates))
	}

	// Verify enriched rate item.
	rate := snapshot.Rates[0]

	// Verify bank code uppercase normalization.
	if rate.BankCode != "NBE" {
		// Error.
		t.Errorf("expected bankCode 'NBE', got '%s'", rate.BankCode)
	}

	// Verify official Arabic name from catalog.
	if rate.NameAr != "البنك الأهلي المصري" {
		// Error.
		t.Errorf("expected official Arabic name, got '%s'", rate.NameAr)
	}

	// Verify brand logo URL from catalog.
	if rate.Logo == "" {
		// Error.
		t.Errorf("expected logo URL to be populated from catalog")
	}

	// Verify downward trend calculation: 49.50 < 49.60.
	if rate.Trend != domain.TrendDown {
		// Error.
		t.Errorf("expected trend 'down' when rate decreased, got '%s'", rate.Trend)
	}

	// Verify availability record.
	if len(snapshot.Availability) != 1 || snapshot.Availability[0].Status != "current" {
		// Error.
		t.Errorf("expected availability record with status 'current'")
	}
}
