// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for JSON serialization, HTTP headers, and time formatting.
import (
	// encoding/json serializes structured error models into standard JSON strings.
	"encoding/json"
	// fmt formats fallback trace identifiers.
	"fmt"
	// math/rand generates random trace suffixes.
	"math/rand"
	// net/http provides HTTP request and response writing primitives.
	"net/http"
	// time formats timestamps according to ISO 8601 UTC representation.
	"time"
)

// ErrorDetail contains the structured code, message, and diagnostic details of an API error.
type ErrorDetail struct {
	// Code represents the machine-readable error token (e.g. "UNAUTHORIZED", "INVALID_VALIDATION", "NOT_FOUND").
	Code string `json:"code"`
	// Message contains human-readable explanation of why the request failed.
	Message string `json:"message"`
	// Details optionally provides structured diagnostic information or field-level validation errors.
	Details any `json:"details,omitempty"`
}

// ErrorResponse represents the standardized public error envelope returned by the gateway.
type ErrorResponse struct {
	// Error contains the nested machine error code and message.
	Error ErrorDetail `json:"error"`
	// TraceID records the distributed request tracing identifier.
	TraceID string `json:"traceId"`
	// Timestamp records the ISO 8601 UTC moment when the error was synthesized.
	Timestamp string `json:"timestamp"`
}

// RespondError serializes and transmits an ErrorResponse matching the platform error schema.
func RespondError(w http.ResponseWriter, status int, code, message string, details any) {
	// Set Content-Type header to UTF-8 encoded application/json.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Write the HTTP response status code header.
	w.WriteHeader(status)

	// Extract or generate trace ID.
	traceID := w.Header().Get("x-trace-id")
	// If missing, generate fallback trace.
	if traceID == "" {
		traceID = fmt.Sprintf("trace-%d-%06d", time.Now().UnixMilli(), rand.Intn(1000000))
		w.Header().Set("x-trace-id", traceID)
	}

	// Construct the error response envelope matching errorEnvelopeSchema.
	errResp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		TraceID:   traceID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Encode and write the JSON payload directly to the HTTP response stream.
	_ = json.NewEncoder(w).Encode(errResp)
}

// RespondJSON serializes and transmits an arbitrary successful data payload with the given status code.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	// Set Content-Type header to UTF-8 encoded application/json.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Write HTTP status code header.
	w.WriteHeader(status)
	// Encode data payload into response stream.
	_ = json.NewEncoder(w).Encode(data)
}
