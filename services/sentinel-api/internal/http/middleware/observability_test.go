package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stellar-account-inspector/sentinel-api/internal/observability"
)

func TestRequestIDAndHTTPObservation(t *testing.T) {
	metrics := observability.NewMetrics()
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := RequestID(ObserveHTTP(logger, metrics, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		request.Pattern = "GET /observed"
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte("ok"))
	})))

	request := httptest.NewRequest(http.MethodGet, "/observed", nil)
	request.Header.Set("X-Request-ID", "valid-request-123")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Header().Get("X-Request-ID") != "valid-request-123" {
		t.Fatalf("request ID = %q", recorder.Header().Get("X-Request-ID"))
	}
	for _, expected := range []string{`"msg":"http_request"`, `"request_id":"valid-request-123"`, `"status":201`, `"route":"GET /observed"`} {
		if !strings.Contains(logs.String(), expected) {
			t.Fatalf("log does not contain %s: %s", expected, logs.String())
		}
	}

	metricsResponse := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(metricsResponse, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsResponse.Body.String()
	if !strings.Contains(body, `sentinel_http_requests_total{method="GET",route="GET /observed",status="201"} 1`) {
		t.Fatalf("HTTP metric missing from exposition: %s", body)
	}
}

func TestRequestIDReplacesInvalidValue(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", "bad value with spaces")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if value := recorder.Header().Get("X-Request-ID"); !requestIDPattern.MatchString(value) || value == "bad value with spaces" {
		t.Fatalf("generated request ID = %q", value)
	}
}
