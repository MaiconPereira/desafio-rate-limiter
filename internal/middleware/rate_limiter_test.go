package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MaiconPereira/desafio-rate-limit/internal/limiter"
)

func TestMiddlewareReturns429WithExpectedBody(t *testing.T) {
	rl := limiter.New(limiter.NewMemoryStore(), limiter.Config{
		IPRequestsPerSecond:    1,
		TokenRequestsPerSecond: 1,
		BlockDuration:          time.Minute,
	})
	middleware := NewRateLimiter(rl)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"

	handler.ServeHTTP(httptest.NewRecorder(), req)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, recorder.Code)
	}
	if recorder.Body.String() != blockedMessage {
		t.Fatalf("expected body %q, got %q", blockedMessage, recorder.Body.String())
	}
}

func TestMiddlewareUsesTokenBeforeIP(t *testing.T) {
	rl := limiter.New(limiter.NewMemoryStore(), limiter.Config{
		IPRequestsPerSecond:    1,
		TokenRequestsPerSecond: 2,
		BlockDuration:          time.Minute,
	})
	middleware := NewRateLimiter(rl)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		req.Header.Set("API_KEY", "premium")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected token request %d to be accepted, got status %d", i+1, recorder.Code)
		}
	}
}
