// Package storage_test provides unit tests for the local filesystem disk persistence adapter.
package storage_test

// Import standard library packages for testing, temp directories, concurrency, and serialization.
import (
	// os provides temporary directory removal utilities.
	"os"
	// sync provides wait groups for concurrent goroutine coordination.
	"sync"
	// testing provides Go unit test primitives.
	"testing"

	// storage points to the disk adapter package under test.
	"bkam-api/internal/adapter/storage"
)

// TestDiskAdapterSnapshotSaveAndLoad verifies that snapshot envelopes are saved atomically and read accurately.
func TestDiskAdapterSnapshotSaveAndLoad(t *testing.T) {
	// Create temporary directory for isolated test execution.
	tmpDir, err := os.MkdirTemp("", "disk-adapter-test-*")
	// Fail test if temporary directory creation failed.
	if err != nil {
		// Log fatal error.
		t.Fatalf("failed to create temp dir: %v", err)
	}
	// Defer cleanup of temporary directory when test concludes.
	defer os.RemoveAll(tmpDir)

	// Initialize disk adapter with temporary directory.
	adapter, err := storage.NewDiskAdapter(tmpDir)
	// Verify initialization succeeds without error.
	if err != nil {
		// Log fatal error.
		t.Fatalf("failed to initialize disk adapter: %v", err)
	}

	// Define sample snapshot data map.
	sampleData := map[string]any{
		// Store test domain key.
		"domain": "rates",
		// Store test count value.
		"count": 28,
		// Store active status flag.
		"status": "current",
	}

	// Execute atomic snapshot write to disk.
	if err := adapter.SaveSnapshot("rates", sampleData); err != nil {
		// Fail test if saving snapshot failed.
		t.Fatalf("failed to save snapshot: %v", err)
	}

	// Prepare receiver map for loading saved snapshot.
	var loadedData map[string]any
	// Read snapshot back from disk.
	if err := adapter.LoadSnapshot("rates", &loadedData); err != nil {
		// Fail test if loading snapshot failed.
		t.Fatalf("failed to load snapshot: %v", err)
	}

	// Assert loaded domain matches expected value.
	if loadedData["domain"] != "rates" {
		// Report mismatch error.
		t.Errorf("expected domain 'rates', got '%v'", loadedData["domain"])
	}

	// Assert loaded status matches expected value.
	if loadedData["status"] != "current" {
		// Report mismatch error.
		t.Errorf("expected status 'current', got '%v'", loadedData["status"])
	}
}

// TestDiskAdapterConcurrentOperations verifies thread safety during concurrent reads and writes.
func TestDiskAdapterConcurrentOperations(t *testing.T) {
	// Create temporary directory for concurrency test.
	tmpDir, err := os.MkdirTemp("", "disk-adapter-concurrency-*")
	// Verify directory creation.
	if err != nil {
		// Log fatal error.
		t.Fatalf("failed to create temp dir: %v", err)
	}
	// Defer removal of temp dir.
	defer os.RemoveAll(tmpDir)

	// Construct disk adapter.
	adapter, err := storage.NewDiskAdapter(tmpDir)
	// Verify no construction errors.
	if err != nil {
		// Log fatal error.
		t.Fatalf("failed to create adapter: %v", err)
	}

	// Initialize wait group for 20 concurrent goroutines.
	var wg sync.WaitGroup
	// Number of concurrent operations.
	workers := 20

	// Launch concurrent writers.
	for i := 0; i < workers; i++ {
		// Increment wait group counter.
		wg.Add(1)
		// Capture iteration index.
		idx := i
		// Launch goroutine worker.
		go func() {
			// Decrement wait group on exit.
			defer wg.Done()
			// Construct iteration payload.
			payload := map[string]int{"iteration": idx}
			// Write snapshot to disk.
			_ = adapter.SaveSnapshot("gold", payload)
		}()
	}

	// Wait for all concurrent operations to conclude.
	wg.Wait()

	// Verify the final snapshot is valid and readable without file corruption.
	var finalPayload map[string]int
	// Attempt loading gold snapshot.
	if err := adapter.LoadSnapshot("gold", &finalPayload); err != nil {
		// Report load error.
		t.Fatalf("failed to read snapshot after concurrent writes: %v", err)
	}
}
