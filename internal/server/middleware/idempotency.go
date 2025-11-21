package middleware

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

// IdempotencyMiddleware ensures requests with the same Idempotency-Key return the same response
func IdempotencyMiddleware() func(http.Handler) http.Handler {
	// Simple thread-safe in-memory cache
	cache := &sync.Map{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to non-idempotent methods
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Get Idempotency-Key header
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if already processed
			if cached, ok := cache.Load(key); ok {
				// Check if it's a completed response or still processing
				if resp, ok := cached.(*cachedResponse); ok {
					w.WriteHeader(resp.status)
					w.Write(resp.body)
					return
				}
				// Still processing, wait
				processing := cached.(*processingMarker)
				<-processing.done
				// Now get the actual response
				if cached, ok := cache.Load(key); ok {
					resp := cached.(*cachedResponse)
					w.WriteHeader(resp.status)
					w.Write(resp.body)
					return
				}
			}

			// Use LoadOrStore to atomically mark as processing
			marker := &processingMarker{done: make(chan struct{})}
			actual, loaded := cache.LoadOrStore(key, marker)

			if loaded {
				// Another goroutine is already processing this key
				processing := actual.(*processingMarker)
				<-processing.done

				// Get the cached response
				if cached, ok := cache.Load(key); ok {
					resp := cached.(*cachedResponse)
					w.WriteHeader(resp.status)
					w.Write(resp.body)
					return
				}
			}

			// This goroutine won the race - process the request
			rec := &recorder{ResponseWriter: w, body: &bytes.Buffer{}, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			// Cache only successful responses (2xx)
			if rec.status >= 200 && rec.status < 300 {
				cache.Store(key, &cachedResponse{
					status: rec.status,
					body:   rec.body.Bytes(),
					time:   time.Now(),
				})
			} else {
				// Remove marker for error responses (don't cache errors)
				cache.Delete(key)
			}

			// Signal that processing is complete
			close(marker.done)
		})
	}
}

// processingMarker indicates a request is currently being processed
type processingMarker struct {
	done chan struct{}
}

// cachedResponse stores an HTTP response
type cachedResponse struct {
	status int
	body   []byte
	time   time.Time
}

// recorder captures status code and response body
type recorder struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
