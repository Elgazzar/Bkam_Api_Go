// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for HTTP interfaces and timestamps.
import (
	// net/http provides response and request primitives.
	"net/http"
	// time provides timestamp generation.
	"time"

	// engine provides the LiveSnapshotCoordinator providing in-memory quotes.
	"bkam-api/internal/engine"
)

// HealthHandler manages health assessment endpoints for the API gateway.
type HealthHandler struct {
	// coordinator supplies in-memory live rates snapshots to evaluate cache freshness.
	coordinator *engine.LiveSnapshotCoordinator
}

// NewHealthHandler constructs a new HealthHandler instance.
func NewHealthHandler(coordinator *engine.LiveSnapshotCoordinator) *HealthHandler {
	// Return handler struct.
	return &HealthHandler{
		// Assign coordinator.
		coordinator: coordinator,
	}
}

// GetHealth handles GET /v1/health, evaluating runtime operational state.
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	// Default system status to "ok".
	status := "ok"

	// Inspect coordinator live rates snapshot if coordinator exists.
	if h.coordinator != nil {
		// Retrieve current rates snapshot from memory.
		if snap, exists := h.coordinator.GetRatesSnapshot(); exists && snap != nil {
			// If freshness indicates partial or stale data.
			if snap.Freshness == "partial" || snap.Freshness == "stale" {
				// Mark system health degraded.
				status = "degraded"
			}
		}
	}

	// Transmit JSON health response payload.
	RespondJSON(w, http.StatusOK, map[string]any{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
