// Package postgres_test provides unit tests for PostgreSQL rates processor and storage.
package postgres_test

import (
	"testing"

	"bkam-api/internal/domain"
	"bkam-api/internal/storage/postgres"
)

// TestRatesProcessorEnrichment verifies that catalog metadata (Arabic/English names, logos) are enriched into rates.
func TestRatesProcessorEnrichment(t *testing.T) {
	processor := postgres.NewRatesProcessor()

	oldBuy := 49.60
	rawDocs := []postgres.RawBankDocument{
		{
			BankID: "nbe",
			Status: "SUCCESS",
			Rates: map[string]postgres.RawBankRateInfo{
				"USD": {
					Buy:    49.50,
					Sell:   49.60,
					OldBuy: &oldBuy,
				},
			},
		},
	}

	snapshot := processor.ProcessRawDocuments(rawDocs, "test-trace-123")
	if snapshot == nil {
		t.Fatalf("expected snapshot to be returned")
	}

	if len(snapshot.Rates) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(snapshot.Rates))
	}

	rate := snapshot.Rates[0]
	if rate.BankCode != "NBE" {
		t.Errorf("expected bankCode 'NBE', got '%s'", rate.BankCode)
	}

	if rate.NameAr != "البنك الأهلي المصري" {
		t.Errorf("expected official Arabic name, got '%s'", rate.NameAr)
	}

	if rate.Logo == "" {
		t.Errorf("expected logo URL to be populated from catalog")
	}

	if rate.Trend != domain.TrendDown {
		t.Errorf("expected trend 'down' when rate decreased, got '%s'", rate.Trend)
	}

	if len(snapshot.Availability) != 1 || snapshot.Availability[0].Status != "current" {
		t.Errorf("expected availability record with status 'current'")
	}
}

// TestRatesProcessorBankCodeFallback verifies that BankCode field is recognized if BankID is empty.
func TestRatesProcessorBankCodeFallback(t *testing.T) {
	processor := postgres.NewRatesProcessor()

	rawDocs := []postgres.RawBankDocument{
		{
			BankCode: "CIB",
			Status:   "SUCCESS",
			Rates: map[string]postgres.RawBankRateInfo{
				"USD": {
					Buy:  48.50,
					Sell: 48.60,
				},
			},
		},
	}

	snapshot := processor.ProcessRawDocuments(rawDocs, "trace-cib")
	if len(snapshot.Rates) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(snapshot.Rates))
	}
	if snapshot.Rates[0].BankCode != "CIB" {
		t.Errorf("expected CIB, got %s", snapshot.Rates[0].BankCode)
	}
}
