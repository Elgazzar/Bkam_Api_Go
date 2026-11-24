// Package engine_test provides unit tests for the LiveSnapshotCoordinator engine.
package engine_test

// Import standard library packages for temporary filesystem, concurrency, and testing.
import (
	// os provides temporary directory removal utilities.
	"os"
	// sync provides WaitGroups for concurrent tests.
	"sync"
	// testing provides Go test primitives.
	"testing"
	// time provides temporal fixtures.
	"time"

	// storage provides the disk adapter for persistence testing.
	"bkam-api/internal/adapter/storage"
	// domain defines core financial models.
	"bkam-api/internal/domain"
	// engine provides the LiveSnapshotCoordinator under test.
	"bkam-api/internal/engine"
)

// TestLiveSnapshotCoordinatorThreadSafety verifies zero race conditions during high-concurrency reads and writes.
func TestLiveSnapshotCoordinatorThreadSafety(t *testing.T) {
	// Create temporary directory for disk persistence.
	tmpDir, err := os.MkdirTemp("", "coordinator-test-*")
	// Verify directory creation.
	if err != nil {
		// Log error.
		t.Fatalf("failed to create temp dir: %v", err)
	}
	// Defer cleanup.
	defer os.RemoveAll(tmpDir)

	// Construct disk adapter.
	disk, err := storage.NewDiskAdapter(tmpDir)
	// Verify disk adapter.
	if err != nil {
		// Log error.
		t.Fatalf("failed to create disk adapter: %v", err)
	}

	// Construct coordinator.
	coord := engine.NewLiveSnapshotCoordinator(disk)

	// Concurrency parameters.
	var wg sync.WaitGroup
	// Workers count.
	workers := 40

	// Launch concurrent writers and readers.
	for i := 0; i < workers; i++ {
		// Increment WaitGroup.
		wg.Add(1)
		// Worker iteration.
		idx := i
		// Launch goroutine.
		go func() {
			// Done on return.
			defer wg.Done()

			// Alternate operations based on index.
			if idx%2 == 0 {
				// Update gold snapshot.
				coord.UpdateSnapshot("gold", &domain.GoldSnapshot{
					Source:        "isagha",
					Currency:      "EGP",
					Freshness:     "current",
					SourceTraceID: "test-trace",
					UpdatedAt:     time.Now().UTC(),
					Rates: map[string]domain.GoldKaratPrice{
						"21k": {Buy: 3500.0, Sell: 3480.0, Trend: domain.TrendUp},
					},
				})
			} else {
				// Read gold snapshot.
				snap, exists := coord.GetGoldSnapshot()
				// If snapshot exists, verify currency field.
				if exists && snap != nil && snap.Currency != "EGP" {
					// Report assertion error.
					t.Errorf("expected currency EGP, got %s", snap.Currency)
				}
			}
		}()
	}

	// Wait for all concurrent routines to finish.
	wg.Wait()
}

// TestLiveSnapshotCoordinatorColdBootHydration verifies disk hydration during cold restarts.
func TestLiveSnapshotCoordinatorColdBootHydration(t *testing.T) {
	// Create temporary directory.
	tmpDir, err := os.MkdirTemp("", "coordinator-hydrate-*")
	// Verify dir.
	if err != nil {
		// Log error.
		t.Fatalf("failed to create temp dir: %v", err)
	}
	// Cleanup on exit.
	defer os.RemoveAll(tmpDir)

	// Construct initial disk adapter.
	disk1, _ := storage.NewDiskAdapter(tmpDir)
	// Construct initial coordinator.
	coord1 := engine.NewLiveSnapshotCoordinator(disk1)

	// Save rates snapshot into coordinator (triggering async disk save).
	testRates := &domain.RateSnapshot{
		Freshness:   "current",
		RequestedAt: time.Now().UTC(),
		Rates: []domain.CurrencyRate{
			{
				BankCode:     "NBE",
				CurrencyCode: "USD",
				BuyRate:      49.50,
				SellRate:     49.60,
				Trend:        domain.TrendUp,
			},
		},
	}
	// Update snapshot in coordinator 1 and save synchronously to disk.
	coord1.UpdateSnapshot("rates", testRates)
	// Explicitly save to disk for synchronous test assertions.
	if err := disk1.SaveSnapshot("rates", testRates); err != nil {
		// Log error.
		t.Fatalf("failed to save snapshot: %v", err)
	}

	// Simulate complete process restart with fresh coordinator instance.
	disk2, _ := storage.NewDiskAdapter(tmpDir)
	// Construct second coordinator.
	coord2 := engine.NewLiveSnapshotCoordinator(disk2)

	// Execute cold boot disk hydration.
	coord2.HydrateFromDisk()

	// Retrieve hydrated rates snapshot.
	hydrated, exists := coord2.GetRatesSnapshot()
	// Assert rates snapshot exists.
	if !exists || hydrated == nil {
		// Fail test.
		t.Fatalf("expected rates snapshot to be hydrated from disk")
	}

	// Verify rate fields.
	if len(hydrated.Rates) != 1 || hydrated.Rates[0].BankCode != "NBE" {
		// Fail assertion.
		t.Errorf("expected hydrated bank code NBE, got %+v", hydrated.Rates)
	}
}
