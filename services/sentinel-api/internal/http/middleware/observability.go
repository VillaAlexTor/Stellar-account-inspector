package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/observability"
)

type requestIDContextKey struct{}

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get("X-Request-ID")
		if !requestIDPattern.MatchString(requestID) {
			requestID = newRequestID()
		}
		writer.Header().Set("X-Request-ID", requestID)
		contextWithID := context.WithValue(request.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(writer, request.WithContext(contextWithID))
	})
}

func ObserveHTTP(logger *slog.Logger, metrics *observability.Metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		finishMetrics := metrics.HTTPStarted()
		recorder := &responseRecorder{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(recorder, request)

		route := request.Pattern
		if route == "" {
			route = "unmatched"
		}
		duration := time.Since(startedAt)
		finishMetrics(request.Method, route, recorder.status, duration)
		logger.InfoContext(request.Context(), "http_request",
			"request_id", RequestIDFromContext(request.Context()),
			"method", request.Method,
			"route", route,
			"status", recorder.status,
			"bytes", recorder.bytes,
			"duration_ms", float64(duration.Microseconds())/1000,
		)
	})
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

type responseRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (recorder *responseRecorder) WriteHeader(status int) {
	if recorder.wroteHeader {
		return
	}
	recorder.wroteHeader = true
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *responseRecorder) Write(payload []byte) (int, error) {
	if !recorder.wroteHeader {
		recorder.WriteHeader(http.StatusOK)
	}
	written, err := recorder.ResponseWriter.Write(payload)
	recorder.bytes += written
	return written, err
}

func (recorder *responseRecorder) Flush() {
	if flusher, ok := recorder.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func newRequestID() string {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err == nil {
		return hex.EncodeToString(value)
	}
	return hex.EncodeToString([]byte(time.Now().UTC().Format("150405.000000")))
}
