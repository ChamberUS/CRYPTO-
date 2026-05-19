package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test that rate limiting middleware blocks excessive requests on payment paths.
func TestRateLimitMiddleware(t *testing.T) {
	alias := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/byx/payments/v1/payment_requests", nil)
	handler := newRateLimitMiddleware(5, 1)(alias)
	tripped := false
	for i := 0; i < 20; i++ {
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code == http.StatusTooManyRequests {
			tripped = true
			break
		}
	}
	if !tripped {
		t.Fatalf("expected rate limiter to trigger 429")
	}
}
