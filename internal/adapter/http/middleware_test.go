// Package http_test provides integration tests for inbound HTTP handlers and routing.
package http_test

// Import standard library packages for HTTP testing.
import (
	// net/http provides request primitives.
	"net/http"
	// net/http/httptest provides response recorders.
	"net/http/httptest"
	// testing provides Go testing fixtures.
	"testing"

	// adapterhttp provides HTTP middleware.
	adapterhttp "bkam-api/internal/adapter/http"
)

// TestTraceMiddleware verifies injection and preservation of distributed trace identifiers.
func TestTraceMiddleware(t *testing.T) {
	// Dummy downstream handler writing 200 OK.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap handler in TraceMiddleware.
	wrapped := adapterhttp.TraceMiddleware(handler)

	// Case 1: Request with custom trace ID.
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("x-trace-id", "custom-trace-12345")
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)
	if rec1.Header().Get("x-trace-id") != "custom-trace-12345" {
		t.Errorf("expected preserved trace ID custom-trace-12345, got %s", rec1.Header().Get("x-trace-id"))
	}

	// Case 2: Request without trace ID (must auto-generate).
	req2, _ := http.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Header().Get("x-trace-id") == "" {
		t.Errorf("expected auto-generated trace ID header on response")
	}
}

// TestAuthMiddleware verifies rejection of unauthorized requests and exemption of public paths.
func TestAuthMiddleware(t *testing.T) {
	// Secret key for testing.
	testSecret := "super-secret-key-123"
	// Dummy downstream handler.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Wrap in AuthMiddleware.
	authMw := adapterhttp.AuthMiddleware(testSecret)
	wrapped := authMw(handler)

	// Case 1: Unauthorized request without header or query key -> 401.
	req1, _ := http.NewRequest("GET", "/v1/rates", nil)
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 for missing key, got %d", rec1.Code)
	}

	// Case 2: Request with valid header x-api-key -> 200.
	req2, _ := http.NewRequest("GET", "/v1/rates", nil)
	req2.Header.Set("x-api-key", testSecret)
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for valid header key, got %d", rec2.Code)
	}

	// Case 3: Request with valid query parameter ?apiKey=secret -> 200.
	req3, _ := http.NewRequest("GET", "/v1/rates?apiKey="+testSecret, nil)
	rec3 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for valid query key, got %d", rec3.Code)
	}

	// Case 4: Public route exemption (/v1/health) without key -> 200.
	req4, _ := http.NewRequest("GET", "/v1/health", nil)
	rec4 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for exempt route /v1/health, got %d", rec4.Code)
	}

	// Case 5: Public route exemption (/v1/version) without key -> 200.
	req5, _ := http.NewRequest("GET", "/v1/version", nil)
	rec5 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec5, req5)
	if rec5.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for exempt route /v1/version, got %d", rec5.Code)
	}
}

// TestETagMiddleware verifies generation of ETag headers and 304 Not Modified revalidation.
func TestETagMiddleware(t *testing.T) {
	// Downstream handler.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"hello world"}`))
	})

	// Wrap in ETagMiddleware.
	wrapped := adapterhttp.ETagMiddleware(handler)

	// Case 1: Initial request receives ETag header.
	req1, _ := http.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec1.Code)
	}
	etag := rec1.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected ETag header on response")
	}

	// Case 2: Subsequent request with matching If-None-Match receives 304 Not Modified.
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotModified {
		t.Errorf("expected HTTP 304 Not Modified, got %d", rec2.Code)
	}
	if rec2.Body.Len() > 0 {
		t.Errorf("expected empty body for 304 Not Modified, got %d bytes", rec2.Body.Len())
	}
}
