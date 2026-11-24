// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for cryptographic hashing, context, formatting, and HTTP primitives.
import (
	// bytes provides in-memory buffer capturing response payloads.
	"bytes"
	// crypto/sha1 generates deterministic digest hashes for HTTP ETags.
	"crypto/sha1"
	// encoding/hex formats byte hashes as lowercase hexadecimal strings.
	"encoding/hex"
	// fmt formats error and trace strings.
	"fmt"
	// math/rand provides pseudo-random suffixes for trace generation.
	"math/rand"
	// net/http provides HTTP handler and middleware types.
	"net/http"
	// strings provides path normalization and prefix checks.
	"strings"
	// time provides timestamp generation for traces.
	"time"
)

// TraceMiddleware ensures every incoming request has an x-trace-id distributed tracing header.
func TraceMiddleware(next http.Handler) http.Handler {
	// Return wrapped handler.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read incoming x-trace-id header.
		traceID := strings.TrimSpace(r.Header.Get("x-trace-id"))
		// If missing, check X-Trace-Id capital casing.
		if traceID == "" {
			traceID = strings.TrimSpace(r.Header.Get("X-Trace-Id"))
		}

		// If still missing, generate deterministic unique trace ID.
		if traceID == "" {
			// Generate trace ID combining millisecond timestamp and random integer.
			traceID = fmt.Sprintf("trace-%d-%06d", time.Now().UnixMilli(), rand.Intn(1000000))
		}

		// Inject trace ID into outgoing response header.
		w.Header().Set("x-trace-id", traceID)

		// Delegate execution to downstream handler.
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates x-api-key header or apiKey query parameter against configured platform secret.
func AuthMiddleware(expectedKey string) func(http.Handler) http.Handler {
	// Return middleware constructor closure.
	return func(next http.Handler) http.Handler {
		// Return wrapped handler.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Normalize request path for exemption checks.
			path := strings.ToLower(strings.TrimSpace(r.URL.Path))

			// Exempt health and version discovery endpoints from authentication requirements.
			if path == "/v1/health" || path == "/v1/version" || path == "/version" {
				// Proceed without auth check.
				next.ServeHTTP(w, r)
				// Return.
				return
			}

			// If no expected key is configured (e.g. offline dev mode), bypass check.
			if expectedKey == "" {
				// Proceed.
				next.ServeHTTP(w, r)
				// Return.
				return
			}

			// Extract API key from x-api-key header.
			apiKey := strings.TrimSpace(r.Header.Get("x-api-key"))
			// Fallback to X-Api-Key casing.
			if apiKey == "" {
				apiKey = strings.TrimSpace(r.Header.Get("X-Api-Key"))
			}

			// If header missing, check query parameters apiKey or apikey.
			if apiKey == "" {
				apiKey = strings.TrimSpace(r.URL.Query().Get("apiKey"))
			}
			if apiKey == "" {
				apiKey = strings.TrimSpace(r.URL.Query().Get("apikey"))
			}

			// Compare provided key with expected platform API key.
			if apiKey != expectedKey {
				// Reject unauthorized request with HTTP 401 UNAUTHORIZED envelope.
				RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid API key.", nil)
				// Halt pipeline.
				return
			}

			// Authorized: continue downstream.
			next.ServeHTTP(w, r)
		})
	}
}

// responseBuffer wraps http.ResponseWriter to intercept response body and status code for ETag calculation.
type responseBuffer struct {
	// ResponseWriter is the underlying response writer.
	http.ResponseWriter
	// body stores the buffered response payload.
	body *bytes.Buffer
	// status stores the written HTTP status code.
	status int
	// headerWritten indicates whether WriteHeader was invoked.
	headerWritten bool
}

// Header returns the response headers map.
func (rb *responseBuffer) Header() http.Header {
	// Return underlying headers.
	return rb.ResponseWriter.Header()
}

// Write captures data into in-memory buffer.
func (rb *responseBuffer) Write(b []byte) (int, error) {
	// If WriteHeader not yet called, default to 200 OK.
	if !rb.headerWritten {
		rb.status = http.StatusOK
		rb.headerWritten = true
	}
	// Write bytes into buffer.
	return rb.body.Write(b)
}

// WriteHeader records HTTP status code.
func (rb *responseBuffer) WriteHeader(statusCode int) {
	// If header not yet written.
	if !rb.headerWritten {
		rb.status = statusCode
		rb.headerWritten = true
	}
}

// ETagMiddleware computes deterministic SHA-1 ETag hashes on JSON responses and validates If-None-Match headers.
func ETagMiddleware(next http.Handler) http.Handler {
	// Return wrapped handler.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only evaluate ETags for idempotent GET and HEAD requests.
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			// Bypass buffering for POST/PUT/DELETE.
			next.ServeHTTP(w, r)
			// Return.
			return
		}

		// Allocate response buffer.
		buf := &responseBuffer{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			status:         http.StatusOK,
		}

		// Execute downstream handler into buffer.
		next.ServeHTTP(buf, r)

		// Only compute ETags for successful 200 OK responses.
		if buf.status == http.StatusOK && buf.body.Len() > 0 {
			// Compute SHA-1 hash of buffered response payload.
			h := sha1.New()
			// Write body bytes into hasher.
			h.Write(buf.body.Bytes())
			// Encode hash as hexadecimal string.
			hashHex := hex.EncodeToString(h.Sum(nil))
			// Construct standard weak ETag string.
			etag := fmt.Sprintf(`W/"%s"`, hashHex)

			// Set ETag header on response.
			w.Header().Set("ETag", etag)

			// Check client If-None-Match request header.
			clientETag := strings.TrimSpace(r.Header.Get("If-None-Match"))
			// Compare ETag match (supporting exact or weak match).
			if clientETag != "" && (clientETag == etag || clientETag == fmt.Sprintf(`"%s"`, hashHex) || clientETag == "*") {
				// Revalidation match: return 304 Not Modified with zero content.
				w.WriteHeader(http.StatusNotModified)
				// Finish handler without writing body.
				return
			}
		}

		// Flush recorded status code to underlying writer.
		w.WriteHeader(buf.status)
		// Flush recorded body bytes to underlying writer.
		w.Write(buf.body.Bytes())
	})
}
