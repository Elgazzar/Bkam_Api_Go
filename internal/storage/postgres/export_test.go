// Package postgres exposes internal helpers for unit testing.
package postgres

import "time"

// ExportDispatchToCoordinator exposes dispatchToCoordinator for unit testing.
func ExportDispatchToCoordinator(s *Storage, domainName, subKey string, payload []byte, updatedAt time.Time) {
	s.dispatchToCoordinator(domainName, subKey, payload, updatedAt)
}
