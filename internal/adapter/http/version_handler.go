// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for HTTP interfaces and time calculation.
import (
	// net/http provides response and request primitives.
	"net/http"
	// runtime provides compiler and Go execution environment metadata.
	"runtime"
	// time provides timestamp and uptime calculation.
	"time"
)

// VersionHandler manages version discovery endpoints (/v1/version and /version).
type VersionHandler struct {
	// startTime records the application process startup timestamp.
	startTime time.Time
	// environment indicates runtime stage ("production", "development", "test").
	environment string
}

// NewVersionHandler constructs and returns a new VersionHandler instance.
func NewVersionHandler(environment string) *VersionHandler {
	// Fallback to "production" if environment is empty.
	if environment == "" {
		environment = "production"
	}
	// Return handler struct.
	return &VersionHandler{
		// Record startup timestamp.
		startTime: time.Now().UTC(),
		// Record environment.
		environment: environment,
	}
}

// GetVersion handles GET /v1/version and GET /version, reporting service metadata and uptime.
func (h *VersionHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	// Calculate uptime duration in seconds.
	uptimeSeconds := int(time.Since(h.startTime).Seconds())

	// Transmit metadata response payload matching public gateway contract.
	RespondJSON(w, http.StatusOK, map[string]any{
		// Semantic API version.
		"version": "3.0.0",
		// Application service name.
		"name": "bkam-api",
		// Deployment environment.
		"environment": h.environment,
		// Runtime platform indicator.
		"runtime": "go",
		// Active Go toolchain version.
		"goVersion": runtime.Version(),
		// Host architecture.
		"architecture": runtime.GOARCH,
		// Compatibility string for legacy clients expecting Node version field.
		"nodeVersion": "go (" + runtime.Version() + ") migrated from Node v22.18.0",
		// Uptime elapsed in seconds.
		"uptime": uptimeSeconds,
		// ISO 8601 current timestamp.
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
