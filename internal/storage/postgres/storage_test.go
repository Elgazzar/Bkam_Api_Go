// Package postgres_test provides unit tests for PostgreSQL in-memory cache and dispatchers.
package postgres_test

import (
	"encoding/json"
	"testing"
	"time"

	"bkam-api/internal/engine"
	"bkam-api/internal/storage/postgres"
)

func TestMemoryCacheOperations(t *testing.T) {
	cache := postgres.NewMemoryCache()

	if count := cache.Count(); count != 0 {
		t.Fatalf("expected empty cache, got %d", count)
	}

	now := time.Now().UTC()
	snap := &postgres.LiveSnapshot{
		Domain:    "rates",
		SubKey:    "",
		Payload:   json.RawMessage(`{"CIB": {"bankCode": "CIB"}}`),
		UpdatedAt: now,
	}

	cache.Set(snap)

	if count := cache.Count(); count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	retrieved, ok := cache.Get("rates", "")
	if !ok || retrieved == nil {
		t.Fatalf("expected to retrieve cached snapshot")
	}

	if string(retrieved.Payload) != string(snap.Payload) {
		t.Errorf("payload mismatch: %s vs %s", retrieved.Payload, snap.Payload)
	}

	// Sub-key test
	stockSnap := &postgres.LiveSnapshot{
		Domain:    "stocks",
		SubKey:    "EGX",
		Payload:   json.RawMessage(`{"market": "EGX"}`),
		UpdatedAt: now,
	}
	cache.Set(stockSnap)

	retrievedStock, ok := cache.Get("stocks", "egx")
	if !ok || retrievedStock == nil {
		t.Fatalf("expected to retrieve stocks/egx case-insensitively")
	}

	all := cache.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 items in GetAll, got %d", len(all))
	}
}

func TestStorageDispatchToCoordinator(t *testing.T) {
	coordinator := engine.NewLiveSnapshotCoordinator(nil)
	storage := postgres.NewStorageWithPool(nil, coordinator)

	now := time.Now().UTC()

	// 1. Rates domain dispatch
	ratesPayload := []byte(`{
		"CIB": {
			"bankCode": "CIB",
			"rates": {
				"USD": {"buy": 48.5, "sell": 48.6}
			}
		}
	}`)
	storage.GetCache().Set(&postgres.LiveSnapshot{
		Domain:    "rates",
		SubKey:    "",
		Payload:   ratesPayload,
		UpdatedAt: now,
	})
	// Trigger dispatch
	ratesSnap, ok := coordinator.GetRatesSnapshot()
	if ok || ratesSnap != nil {
		t.Fatalf("expected rates snapshot not yet in coordinator")
	}

	// Now hydrate via simulated row
	simulateHydrate(storage, "rates", "", ratesPayload, now)
	ratesSnap, ok = coordinator.GetRatesSnapshot()
	if !ok || ratesSnap == nil {
		t.Fatalf("expected rates snapshot in coordinator after dispatch")
	}
	if len(ratesSnap.Rates) != 1 || ratesSnap.Rates[0].BankCode != "CIB" {
		t.Errorf("expected CIB rate, got %+v", ratesSnap.Rates)
	}

	// 2. Gold domain dispatch
	goldPayload := []byte(`{
		"lastUpdated": "2026-11-28T00:00:00Z",
		"carats": {
			"24k": {"buy": 4200, "sell": 4250},
			"21k": {"buy": 3675, "sell": 3720}
		},
		"spotUsd": 2650.5
	}`)
	simulateHydrate(storage, "gold", "", goldPayload, now)
	goldSnap, ok := coordinator.GetGoldSnapshot()
	if !ok || goldSnap == nil {
		t.Fatalf("expected gold snapshot in coordinator")
	}
	if len(goldSnap.Rates) != 2 {
		t.Errorf("expected 2 karats, got %d", len(goldSnap.Rates))
	}
	if goldSnap.International == nil || goldSnap.International.Price != 2650.5 {
		t.Errorf("expected international spot price 2650.5, got %+v", goldSnap.International)
	}

	// 3. Gas domain dispatch
	gasPayload := []byte(`{
		"lastUpdated": "2026-11-28T00:00:00Z",
		"fuels": {
			"octane92": 15.25,
			"diesel": 13.50
		}
	}`)
	simulateHydrate(storage, "gas", "", gasPayload, now)
	gasSnap, ok := coordinator.GetGasSnapshot()
	if !ok || gasSnap == nil {
		t.Fatalf("expected gas snapshot in coordinator")
	}
	if len(gasSnap.Rates) != 2 {
		t.Errorf("expected 2 fuel rates, got %d", len(gasSnap.Rates))
	}
	if gasSnap.Rates["octane92"].Price != 15.25 {
		t.Errorf("expected octane92 price 15.25, got %f", gasSnap.Rates["octane92"].Price)
	}

	// 4. Stocks domain dispatch (EGX)
	egxPayload := []byte(`{
		"market": "EGX",
		"lastUpdated": "2026-11-28T00:00:00Z",
		"quotes": [
			{"symbol": "COMI", "price": 85.5, "change": 1.2, "changePercent": 1.4}
		]
	}`)
	simulateHydrate(storage, "stocks", "EGX", egxPayload, now)
	egxSnap, ok := coordinator.GetStockSnapshot("egx")
	if !ok || egxSnap == nil {
		t.Fatalf("expected EGX stock snapshot in coordinator")
	}
	if len(egxSnap.Stocks) != 1 || egxSnap.Stocks["COMI"].Price != 85.5 {
		t.Errorf("expected COMI price 85.5, got %+v", egxSnap.Stocks)
	}

	// 5. Stocks domain dispatch (Tadawul / Saudi)
	tadawulPayload := []byte(`{
		"market": "TADAWUL",
		"lastUpdated": "2026-11-28T00:00:00Z",
		"quotes": [
			{"symbol": "2222", "price": 32.5, "change": 0.5, "changePercent": 1.5}
		]
	}`)
	simulateHydrate(storage, "stocks", "TADAWUL", tadawulPayload, now)
	saudiSnap, ok := coordinator.GetStockSnapshot("saudi")
	if !ok || saudiSnap == nil {
		t.Fatalf("expected Saudi stock snapshot in coordinator")
	}
	if len(saudiSnap.Stocks) != 1 || saudiSnap.Stocks["2222"].Price != 32.5 {
		t.Errorf("expected 2222 stock price 32.5, got %+v", saudiSnap.Stocks)
	}
}

// simulateHydrate exposes internal dispatch logic via cache setting and coordinator update.
func simulateHydrate(s *postgres.Storage, domainName, subKey string, payload []byte, updatedAt time.Time) {
	// Directly call exported or package internal dispatch
	// Storage has memory cache set and dispatchToCoordinator
	s.GetCache().Set(&postgres.LiveSnapshot{
		Domain:    domainName,
		SubKey:    subKey,
		Payload:   payload,
		UpdatedAt: updatedAt,
	})
	// Trigger hydration simulation
	postgres.ExportDispatchToCoordinator(s, domainName, subKey, payload, updatedAt)
}
