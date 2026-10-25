package observability

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const metricNamespace = "publication_platform"

const serviceName = "publication-platform-backend"

type Metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
}

func NewJSONLogger(output io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			switch attr.Key {
			case slog.TimeKey:
				attr.Key = "timestamp"
			case slog.MessageKey:
				attr.Key = "message"
			}
			return attr
		},
	})

	return slog.New(handler).With("service.name", serviceName)
}

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapped, r)
			status := responseStatus(wrapped)
			if isHealthCheck(r) && status < http.StatusBadRequest {
				return
			}

			logger.InfoContext(r.Context(), "http_request",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"route", routePattern(r),
				"status", status,
				"bytes", wrapped.BytesWritten(),
				"duration_ms", float64(time.Since(startedAt).Microseconds())/1000,
			)
		})
	}
}

func isHealthCheck(r *http.Request) bool {
	return r.URL.Path == "/health/live" || r.URL.Path == "/health/ready"
}

func NewMetrics(registry *prometheus.Registry, db *sql.DB) *Metrics {
	metrics := &Metrics{
		registry: registry,
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricNamespace,
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route", "status"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Current number of HTTP requests being served.",
		}),
	}

	registry.MustRegister(metrics.requests, metrics.duration, metrics.inFlight)
	registerDBMetrics(registry, db)
	return metrics
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		m.inFlight.Inc()
		defer m.inFlight.Dec()

		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(wrapped, r)

		labels := []string{r.Method, routePattern(r), strconv.Itoa(responseStatus(wrapped))}
		m.requests.WithLabelValues(labels...).Inc()
		m.duration.WithLabelValues(labels...).Observe(time.Since(startedAt).Seconds())
	})
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

func routePattern(r *http.Request) string {
	pattern := chi.RouteContext(r.Context()).RoutePattern()
	if pattern == "" {
		return "unmatched"
	}
	return pattern
}

func responseStatus(w middleware.WrapResponseWriter) int {
	if w.Status() == 0 {
		return http.StatusOK
	}
	return w.Status()
}

func registerDBMetrics(registry prometheus.Registerer, db *sql.DB) {
	gauges := []struct {
		name string
		help string
		read func(sql.DBStats) float64
	}{
		{name: "connections_open", help: "Number of established database connections.", read: func(s sql.DBStats) float64 { return float64(s.OpenConnections) }},
		{name: "connections_in_use", help: "Number of database connections currently in use.", read: func(s sql.DBStats) float64 { return float64(s.InUse) }},
		{name: "connections_idle", help: "Number of idle database connections.", read: func(s sql.DBStats) float64 { return float64(s.Idle) }},
		{name: "connections_max_open", help: "Maximum number of open database connections.", read: func(s sql.DBStats) float64 { return float64(s.MaxOpenConnections) }},
	}

	for _, gauge := range gauges {
		read := gauge.read
		registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: "db",
			Name:      gauge.name,
			Help:      gauge.help,
		}, func() float64 {
			return read(db.Stats())
		}))
	}

	registry.MustRegister(prometheus.NewCounterFunc(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Subsystem: "db",
		Name:      "connection_wait_total",
		Help:      "Total number of waits for a database connection.",
	}, func() float64 {
		return float64(db.Stats().WaitCount)
	}))
}
