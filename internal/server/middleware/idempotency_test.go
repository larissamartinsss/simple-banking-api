package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdempotencyMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		idempotencyKey string
		requestCount   int
		wantCallCount  int
	}{
		{
			name:           "POST with idempotency key - second request returns cached response",
			method:         "POST",
			idempotencyKey: "test-key-123",
			requestCount:   2,
			wantCallCount:  1, // Handler called only once
		},
		{
			name:           "POST without idempotency key - all requests processed",
			method:         "POST",
			idempotencyKey: "",
			requestCount:   2,
			wantCallCount:  2, // Handler called twice
		},
		{
			name:           "GET request - idempotency not applied",
			method:         "GET",
			idempotencyKey: "test-key-456",
			requestCount:   2,
			wantCallCount:  2, // Handler called twice (GET is naturally idempotent)
		},
		{
			name:           "POST with different keys - all processed",
			method:         "POST",
			idempotencyKey: "", // Will use different keys per request
			requestCount:   2,
			wantCallCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track how many times handler is called
			callCount := 0

			// Create a test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id":1,"status":"created"}`))
			})

			// Wrap with idempotency middleware
			middleware := IdempotencyMiddleware()
			wrappedHandler := middleware(handler)

			// Make multiple requests
			for i := 0; i < tt.requestCount; i++ {
				req := httptest.NewRequest(tt.method, "/test", strings.NewReader(`{"test":"data"}`))
				req.Header.Set("Content-Type", "application/json")

				// Set idempotency key
				if tt.idempotencyKey != "" {
					req.Header.Set("Idempotency-Key", tt.idempotencyKey)
				} else if tt.name == "POST with different keys - all processed" {
					req.Header.Set("Idempotency-Key", "unique-key-"+string(rune(i)))
				}

				rec := httptest.NewRecorder()
				wrappedHandler.ServeHTTP(rec, req)

				// Verify response
				assert.Equal(t, http.StatusCreated, rec.Code)
				body, _ := io.ReadAll(rec.Body)
				assert.Contains(t, string(body), `"id":1`)
			}

			// Verify handler was called expected number of times
			assert.Equal(t, tt.wantCallCount, callCount, "Handler call count mismatch")
		})
	}
}

func TestIdempotencyMiddleware_CachesOnlySuccessfulResponses(t *testing.T) {
	callCount := 0

	// Handler that fails first time, succeeds second time
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"server error"}`))
		} else {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":1,"status":"created"}`))
		}
	})

	middleware := IdempotencyMiddleware()
	wrappedHandler := middleware(handler)

	// First request - fails (500)
	req1 := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test":"data"}`))
	req1.Header.Set("Idempotency-Key", "test-key-error")
	rec1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusInternalServerError, rec1.Code)

	// Second request - should NOT use cache (first was error), processes again and succeeds
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test":"data"}`))
	req2.Header.Set("Idempotency-Key", "test-key-error")
	rec2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusCreated, rec2.Code)

	// Verify handler was called twice (error responses not cached)
	assert.Equal(t, 2, callCount)
}

func TestIdempotencyMiddleware_ReturnsIdenticalResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":123,"amount":100.50}`))
	})

	middleware := IdempotencyMiddleware()
	wrappedHandler := middleware(handler)

	// First request
	req1 := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test":"data"}`))
	req1.Header.Set("Idempotency-Key", "same-response-test")
	rec1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec1, req1)

	// Second request with same key
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test":"data"}`))
	req2.Header.Set("Idempotency-Key", "same-response-test")
	rec2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec2, req2)

	// Both responses should be identical
	assert.Equal(t, rec1.Code, rec2.Code, "Status codes should match")
	assert.Equal(t, rec1.Body.String(), rec2.Body.String(), "Response bodies should match")
}
