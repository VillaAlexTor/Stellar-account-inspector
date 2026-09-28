package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testAccount = "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

func TestRateLimiterLimitsByIP(t *testing.T) {
	limiter := NewRateLimiter(RateLimitOptions{Window: time.Minute, IPLimit: 2, AccountLimit: 20})
	limiter.now = func() time.Time { return time.Date(2026, time.September, 27, 12, 0, 0, 0, time.UTC) }
	handler := limiter.Handler(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	for attempt := 1; attempt <= 3; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/session", nil)
		request.RemoteAddr = "192.0.2.10:4321"
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if attempt <= 2 && recorder.Code != http.StatusNoContent {
			t.Fatalf("attempt %d status = %d", attempt, recorder.Code)
		}
		if attempt == 3 {
			if recorder.Code != http.StatusTooManyRequests {
				t.Fatalf("limited status = %d", recorder.Code)
			}
			if recorder.Header().Get("Retry-After") == "" {
				t.Fatal("limited response did not include Retry-After")
			}
		}
	}
}

func TestRateLimiterLimitsAccountAcrossIPs(t *testing.T) {
	limiter := NewRateLimiter(RateLimitOptions{Window: time.Minute, IPLimit: 20, AccountLimit: 1})
	handler := limiter.Handler(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	for index, address := range []string{"192.0.2.10:4000", "198.51.100.20:4000"} {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/monitored-accounts/"+testAccount+"/alerts", nil)
		request.RemoteAddr = address
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if index == 0 && recorder.Code != http.StatusNoContent {
			t.Fatalf("first account request status = %d", recorder.Code)
		}
		if index == 1 && recorder.Code != http.StatusTooManyRequests {
			t.Fatalf("second account request status = %d", recorder.Code)
		}
	}
}

func TestRateLimiterOnlyTrustsForwardedForFromConfiguredProxy(t *testing.T) {
	limiter := NewRateLimiter(RateLimitOptions{
		Window:            time.Minute,
		IPLimit:           20,
		AccountLimit:      20,
		TrustedProxyCIDRs: []string{"10.0.0.0/8"},
	})

	untrusted := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	untrusted.RemoteAddr = "192.0.2.5:1234"
	untrusted.Header.Set("X-Forwarded-For", "203.0.113.9")
	if got := limiter.clientIP(untrusted); got != "192.0.2.5" {
		t.Fatalf("untrusted proxy client IP = %q", got)
	}

	trusted := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	trusted.RemoteAddr = "10.1.2.3:1234"
	trusted.Header.Set("X-Forwarded-For", "203.0.113.9, 10.2.3.4")
	if got := limiter.clientIP(trusted); got != "203.0.113.9" {
		t.Fatalf("trusted proxy client IP = %q", got)
	}
}
