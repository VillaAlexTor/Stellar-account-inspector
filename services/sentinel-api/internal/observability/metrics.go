package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry          *prometheus.Registry
	httpRequests      *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
	httpInflight      prometheus.Gauge
	sseConnections    prometheus.Gauge
	authAttempts      *prometheus.CounterVec
	rateLimitRejected *prometheus.CounterVec
	horizonEvents     *prometheus.CounterVec
	alerts            *prometheus.CounterVec
	monitoredSessions prometheus.Gauge
	databaseReady     prometheus.Gauge
	notifications     *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	metrics := &Metrics{
		registry: prometheus.NewRegistry(),
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sentinel", Name: "http_requests_total", Help: "Solicitudes HTTP procesadas por Sentinel.",
		}, []string{"method", "route", "status"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "sentinel", Name: "http_request_duration_seconds", Help: "Duración de solicitudes HTTP.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		httpInflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sentinel", Name: "http_requests_in_flight", Help: "Solicitudes HTTP actualmente en curso.",
		}),
		sseConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sentinel", Name: "sse_connections", Help: "Conexiones SSE activas hacia navegadores.",
		}),
		authAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sentinel", Name: "auth_attempts_total", Help: "Intentos de autenticación por resultado.",
		}, []string{"result"}),
		rateLimitRejected: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sentinel", Name: "rate_limit_rejections_total", Help: "Solicitudes rechazadas por límite y alcance.",
		}, []string{"scope"}),
		horizonEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sentinel", Name: "horizon_stream_events_total", Help: "Eventos operativos de streams Horizon.",
		}, []string{"event"}),
		alerts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sentinel", Name: "alerts_total", Help: "Alertas generadas por regla y severidad.",
		}, []string{"rule_id", "severity"}),
		monitoredSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sentinel", Name: "monitored_sessions", Help: "Sesiones de monitoreo únicas mantenidas por el proceso.",
		}),
		databaseReady: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "sentinel", Name: "database_ready", Help: "Disponibilidad actual de PostgreSQL (1 disponible, 0 no disponible).",
		}),
		notifications: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sentinel", Name: "notification_deliveries_total", Help: "Entregas de notificaciones por canal y resultado.",
		}, []string{"channel", "result"}),
	}
	metrics.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		metrics.httpRequests,
		metrics.httpDuration,
		metrics.httpInflight,
		metrics.sseConnections,
		metrics.authAttempts,
		metrics.rateLimitRejected,
		metrics.horizonEvents,
		metrics.alerts,
		metrics.monitoredSessions,
		metrics.databaseReady,
		metrics.notifications,
	)
	return metrics
}

func (metrics *Metrics) Notification(channel, result string) {
	metrics.notifications.WithLabelValues(channel, result).Inc()
}

func (metrics *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(metrics.registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

func (metrics *Metrics) HTTPStarted() func(method, route string, status int, duration time.Duration) {
	metrics.httpInflight.Inc()
	return func(method, route string, status int, duration time.Duration) {
		metrics.httpInflight.Dec()
		metrics.httpRequests.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
		metrics.httpDuration.WithLabelValues(method, route).Observe(duration.Seconds())
	}
}

func (metrics *Metrics) SSEConnected() func() {
	metrics.sseConnections.Inc()
	return metrics.sseConnections.Dec
}

func (metrics *Metrics) AuthAttempt(result string) {
	metrics.authAttempts.WithLabelValues(result).Inc()
}

func (metrics *Metrics) RateLimitRejected(scope string) {
	metrics.rateLimitRejected.WithLabelValues(scope).Inc()
}

func (metrics *Metrics) HorizonEvent(event string) {
	metrics.horizonEvents.WithLabelValues(event).Inc()
}

func (metrics *Metrics) Alert(ruleID, severity string) {
	metrics.alerts.WithLabelValues(ruleID, severity).Inc()
}

func (metrics *Metrics) SetMonitoredSessions(count int) {
	metrics.monitoredSessions.Set(float64(count))
}

func (metrics *Metrics) SetDatabaseReady(ready bool) {
	if ready {
		metrics.databaseReady.Set(1)
		return
	}
	metrics.databaseReady.Set(0)
}
