// Package postgres provides PostgreSQL 17 persistence and real-time LISTEN/NOTIFY in-memory cache invalidation.
package postgres

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// LiveSnapshot represents a stored snapshot row from the live_snapshots table.
type LiveSnapshot struct {
	// Domain identifies the market domain (e.g. 'rates', 'gold', 'gas', 'stocks', 'indices').
	Domain string `json:"domain"`
	// SubKey identifies the market or sub-category (e.g. 'EGX', 'TADAWUL', or empty string).
	SubKey string `json:"sub_key"`
	// Payload holds the raw JSONB document.
	Payload json.RawMessage `json:"payload"`
	// UpdatedAt records the database modification timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoryCache provides a high-performance, thread-safe in-memory cache for live snapshots.
type MemoryCache struct {
	mu        sync.RWMutex
	snapshots map[string]*LiveSnapshot
}

// NewMemoryCache constructs an empty in-memory snapshot cache.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		snapshots: make(map[string]*LiveSnapshot),
	}
}

// makeKey computes a canonical composite key for (domain, sub_key).
func (c *MemoryCache) makeKey(domainName, subKey string) string {
	d := strings.ToLower(strings.TrimSpace(domainName))
	s := strings.ToLower(strings.TrimSpace(subKey))
	if s != "" {
		return fmt.Sprintf("%s/%s", d, s)
	}
	return d
}

// Get retrieves a snapshot from memory cache in 0ms without database queries.
func (c *MemoryCache) Get(domainName, subKey string) (*LiveSnapshot, bool) {
	key := c.makeKey(domainName, subKey)
	c.mu.RLock()
	defer c.mu.RUnlock()
	snap, ok := c.snapshots[key]
	return snap, ok
}

// Set stores or updates a snapshot in memory cache.
func (c *MemoryCache) Set(snap *LiveSnapshot) {
	if snap == nil {
		return
	}
	key := c.makeKey(snap.Domain, snap.SubKey)
	c.mu.Lock()
	c.snapshots[key] = snap
	c.mu.Unlock()
}

// GetAll returns a copy of all cached snapshots.
func (c *MemoryCache) GetAll() []*LiveSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make([]*LiveSnapshot, 0, len(c.snapshots))
	for _, snap := range c.snapshots {
		res = append(res, snap)
	}
	return res
}

// Count returns the number of active cached snapshots.
func (c *MemoryCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.snapshots)
}
