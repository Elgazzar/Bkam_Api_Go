// Package engine implements substantive core domain engines adhering to John Ousterhout's Deep Module design philosophy.
package engine

// Import standard library and external packages for concurrency, synchronization, and storage.
import (
	// fmt provides formatted string generation for error messages and cache keys.
	"fmt"
	// strings provides uppercase normalization and whitespace trimming.
	"strings"
	// sync provides RWMutex for thread-safe concurrent reads and writes to memory state.
	"sync"
	// time provides timestamping for memory updates and telemetry.
	"time"

	// singleflight coalesces duplicate concurrent requests onto a single in-flight execution promise.
	"golang.org/x/sync/singleflight"

	// storage provides local filesystem persistence for cold boot hydration.
	"bkam-api/internal/adapter/storage"
	// domain defines core financial models and snapshot envelopes.
	"bkam-api/internal/domain"
)

// LiveSnapshotCoordinator maintains thread-safe in-memory market quotes and manages disk persistence.
type LiveSnapshotCoordinator struct {
	// mu protects concurrent access to the in-memory snapshots map.
	mu sync.RWMutex
	// snapshots stores the latest deserialized domain snapshot envelopes in memory.
	snapshots map[string]any
	// lastUpdated stores the timestamp of the latest in-memory mutation per domain key.
	lastUpdated map[string]time.Time
	// diskAdapter handles local filesystem persistence for zero-read restarts.
	diskAdapter *storage.DiskAdapter
	// flightGroup coalesces concurrent reads during cold cache warming.
	flightGroup singleflight.Group
}

// NewLiveSnapshotCoordinator constructs and initializes a new LiveSnapshotCoordinator.
func NewLiveSnapshotCoordinator(diskAdapter *storage.DiskAdapter) *LiveSnapshotCoordinator {
	// Initialize coordinator struct with allocated maps.
	return &LiveSnapshotCoordinator{
		// Allocate snapshot store map.
		snapshots: make(map[string]any),
		// Allocate last updated timestamp map.
		lastUpdated: make(map[string]time.Time),
		// Store disk persistence adapter reference.
		diskAdapter: diskAdapter,
	}
}

// makeKey generates a canonical composite lookup key from domain and optional market code.
func (c *LiveSnapshotCoordinator) makeKey(domainName string, market ...string) string {
	// Normalize domain string to lowercase.
	normDomain := strings.ToLower(strings.TrimSpace(domainName))
	// If a market code is provided, append it.
	if len(market) > 0 && market[0] != "" {
		// Form composite key e.g. "gcc-stocks/saudi".
		return fmt.Sprintf("%s/%s", normDomain, strings.ToLower(strings.TrimSpace(market[0])))
	}
	// Return standalone domain key.
	return normDomain
}

// UpdateSnapshot updates an asset snapshot in memory and triggers asynchronous disk persistence.
func (c *LiveSnapshotCoordinator) UpdateSnapshot(domainName string, data any, market ...string) {
	// Compute canonical composite lookup key.
	key := c.makeKey(domainName, market...)

	// Acquire exclusive write lock to update memory state safely.
	c.mu.Lock()
	// Store updated snapshot in memory.
	c.snapshots[key] = data
	// Record update timestamp.
	c.lastUpdated[key] = time.Now().UTC()
	// Release write lock.
	c.mu.Unlock()

	// Asynchronously persist snapshot to local disk if disk adapter is available.
	if c.diskAdapter != nil {
		// Launch background worker goroutine for non-blocking disk persistence.
		go func() {
			// Extract optional market sub-identifier.
			m := ""
			// Assign market if present.
			if len(market) > 0 {
				// Set market code.
				m = market[0]
			}
			// Persist snapshot to disk.
			_ = c.diskAdapter.SaveSnapshot(domainName, data, m)
		}()
	}
}

// GetSnapshot retrieves the in-memory snapshot for a domain, returning false if not yet available.
func (c *LiveSnapshotCoordinator) GetSnapshot(domainName string, market ...string) (any, bool) {
	// Compute canonical composite key.
	key := c.makeKey(domainName, market...)

	// Acquire shared read lock.
	c.mu.RLock()
	// Defer release of read lock upon function return.
	defer c.mu.RUnlock()

	// Look up snapshot in memory map.
	data, exists := c.snapshots[key]
	// Return data and existence flag.
	return data, exists
}

// GetRatesSnapshot retrieves the typed RateSnapshot envelope from memory.
func (c *LiveSnapshotCoordinator) GetRatesSnapshot() (*domain.RateSnapshot, bool) {
	// Fetch raw snapshot for "rates" domain.
	raw, exists := c.GetSnapshot("rates")
	// If missing, return false.
	if !exists || raw == nil {
		// Return nil and false.
		return nil, false
	}

	// Cast to *domain.RateSnapshot pointer.
	if typed, ok := raw.(*domain.RateSnapshot); ok {
		// Return typed pointer.
		return typed, true
	}
	// Attempt value cast.
	if typed, ok := raw.(domain.RateSnapshot); ok {
		// Return address of typed value.
		return &typed, true
	}

	// Return nil if cast fails.
	return nil, false
}

// GetGoldSnapshot retrieves the typed GoldSnapshot envelope from memory.
func (c *LiveSnapshotCoordinator) GetGoldSnapshot() (*domain.GoldSnapshot, bool) {
	// Fetch raw snapshot for "gold" domain.
	raw, exists := c.GetSnapshot("gold")
	// If missing, return false.
	if !exists || raw == nil {
		// Return nil and false.
		return nil, false
	}

	// Cast to *domain.GoldSnapshot.
	if typed, ok := raw.(*domain.GoldSnapshot); ok {
		// Return typed pointer.
		return typed, true
	}
	// Attempt value cast.
	if typed, ok := raw.(domain.GoldSnapshot); ok {
		// Return address.
		return &typed, true
	}

	// Cast failure fallback.
	return nil, false
}

// GetGasSnapshot retrieves the typed GasSnapshot envelope from memory.
func (c *LiveSnapshotCoordinator) GetGasSnapshot() (*domain.GasSnapshot, bool) {
	// Fetch raw snapshot for "gas" domain.
	raw, exists := c.GetSnapshot("gas")
	// If missing, return false.
	if !exists || raw == nil {
		// Return nil and false.
		return nil, false
	}

	// Cast to *domain.GasSnapshot.
	if typed, ok := raw.(*domain.GasSnapshot); ok {
		// Return typed pointer.
		return typed, true
	}
	// Attempt value cast.
	if typed, ok := raw.(domain.GasSnapshot); ok {
		// Return address.
		return &typed, true
	}

	// Cast failure fallback.
	return nil, false
}

// GetStockSnapshot retrieves the typed StockSnapshot for a specific market exchange.
func (c *LiveSnapshotCoordinator) GetStockSnapshot(market string) (*domain.StockSnapshot, bool) {
	// Normalize market code.
	m := strings.ToLower(strings.TrimSpace(market))
	// Determine domain name based on market code.
	domainName := "stocks"
	// If not EGX, use "gcc-stocks" domain identifier.
	if m != "egx" && m != "" {
		// Use gcc-stocks domain.
		domainName = "gcc-stocks"
	}

	// Fetch raw snapshot.
	raw, exists := c.GetSnapshot(domainName, m)
	// If missing, return false.
	if !exists || raw == nil {
		// Return nil and false.
		return nil, false
	}

	// Cast to *domain.StockSnapshot.
	if typed, ok := raw.(*domain.StockSnapshot); ok {
		// Return typed pointer.
		return typed, true
	}
	// Attempt value cast.
	if typed, ok := raw.(domain.StockSnapshot); ok {
		// Return address.
		return &typed, true
	}

	// Cast failure fallback.
	return nil, false
}

// GetIndexSnapshot retrieves the typed IndexSnapshot for EGX or regional GCC benchmark indices.
func (c *LiveSnapshotCoordinator) GetIndexSnapshot(market ...string) (*domain.IndexSnapshot, bool) {
	// Determine target market (defaulting to "egx").
	m := "egx"
	// Check if custom market provided.
	if len(market) > 0 && market[0] != "" {
		// Assign market.
		m = strings.ToLower(strings.TrimSpace(market[0]))
	}

	// Fetch raw snapshot.
	raw, exists := c.GetSnapshot("indices", m)
	// If missing, return false.
	if !exists || raw == nil {
		// Return nil and false.
		return nil, false
	}

	// Cast to *domain.IndexSnapshot.
	if typed, ok := raw.(*domain.IndexSnapshot); ok {
		// Return typed pointer.
		return typed, true
	}
	// Attempt value cast.
	if typed, ok := raw.(domain.IndexSnapshot); ok {
		// Return address.
		return &typed, true
	}

	// Cast failure fallback.
	return nil, false
}

// HydrateFromDisk reads saved snapshots from the local filesystem during cold boot.
func (c *LiveSnapshotCoordinator) HydrateFromDisk() {
	// If no disk adapter configured, abort.
	if c.diskAdapter == nil {
		// Return early.
		return
	}

	// Attempt hydrating currency rates snapshot.
	var rates domain.RateSnapshot
	// Load from disk.
	if err := c.diskAdapter.LoadSnapshot("rates", &rates); err == nil && len(rates.Rates) > 0 {
		// Store in memory.
		c.UpdateSnapshot("rates", &rates)
	}

	// Attempt hydrating precious metals gold snapshot.
	var gold domain.GoldSnapshot
	// Load from disk.
	if err := c.diskAdapter.LoadSnapshot("gold", &gold); err == nil && len(gold.Rates) > 0 {
		// Store in memory.
		c.UpdateSnapshot("gold", &gold)
	}

	// Attempt hydrating petroleum fuel gas snapshot.
	var gas domain.GasSnapshot
	// Load from disk.
	if err := c.diskAdapter.LoadSnapshot("gas", &gas); err == nil && len(gas.Rates) > 0 {
		// Store in memory.
		c.UpdateSnapshot("gas", &gas)
	}

	// Attempt hydrating EGX stocks snapshot.
	var egxStocks domain.StockSnapshot
	// Load from disk.
	if err := c.diskAdapter.LoadSnapshot("stocks", &egxStocks, "egx"); err == nil && len(egxStocks.Stocks) > 0 {
		// Store in memory.
		c.UpdateSnapshot("stocks", &egxStocks, "egx")
	}

	// Attempt hydrating EGX benchmark indices snapshot.
	var egxIndices domain.IndexSnapshot
	// Load from disk.
	if err := c.diskAdapter.LoadSnapshot("indices", &egxIndices, "egx"); err == nil && len(egxIndices.Indices) > 0 {
		// Store in memory.
		c.UpdateSnapshot("indices", &egxIndices, "egx")
	}

	// Hydrate regional GCC stock markets.
	gccMarkets := []string{"saudi", "uae", "qatar", "kuwait"}
	// Loop over GCC markets.
	for _, mkt := range gccMarkets {
		// Prepare target struct.
		var gccStocks domain.StockSnapshot
		// Load from disk.
		if err := c.diskAdapter.LoadSnapshot("gcc-stocks", &gccStocks, mkt); err == nil && len(gccStocks.Stocks) > 0 {
			// Store in memory.
			c.UpdateSnapshot("gcc-stocks", &gccStocks, mkt)
		}
	}

	// Hydrate regional benchmark indices for GCC markets.
	regionalIndexMarkets := []string{"saudi", "uae", "qatar"}
	// Loop over regional index markets.
	for _, mkt := range regionalIndexMarkets {
		// Prepare target struct.
		var regIndices domain.IndexSnapshot
		// Load from disk.
		if err := c.diskAdapter.LoadSnapshot("indices", &regIndices, mkt); err == nil && len(regIndices.Indices) > 0 {
			// Store in memory.
			c.UpdateSnapshot("indices", &regIndices, mkt)
		}
	}
}
