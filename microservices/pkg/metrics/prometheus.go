package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all HTTP-related metrics
type Metrics struct {
	requestsTotal     *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
	errorTotal        *prometheus.CounterVec
	activeConnections *prometheus.GaugeVec
}

// Config holds configuration for metrics
type Config struct {
	Namespace string
	Subsystem string
}

// New creates a new metrics instance with Prometheus metrics
func New(cfg Config) *Metrics {
	if cfg.Namespace == "" {
		cfg.Namespace = "http"
	}
	if cfg.Subsystem == "" {
		cfg.Subsystem = "requests"
	}

	return &Metrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"service", "method", "path", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "method", "path", "status"},
		),
		errorTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "errors_total",
				Help:      "Total number of HTTP errors",
			},
			[]string{"service", "method", "path", "status", "error_type"},
		),
		activeConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "active_connections",
				Help:      "Current number of active HTTP connections",
			},
			[]string{"service", "method", "path"},
		),
	}
}

// RecordRequest manually records request metrics
func (m *Metrics) RecordRequest(service, method, path, status string, duration time.Duration) {
	labels := []string{service, method, path, status}

	m.requestsTotal.WithLabelValues(labels...).Inc()
	m.requestDuration.WithLabelValues(labels...).Observe(duration.Seconds())
}

// RecordError records error metrics with additional error type information
func (m *Metrics) RecordError(service, method, path, status, errorType string) {
	labels := []string{service, method, path, status, errorType}
	m.errorTotal.WithLabelValues(labels...).Inc()
}

// SetActiveConnections sets the gauge for active connections
func (m *Metrics) SetActiveConnections(service, method, path string, count float64) {
	labels := []string{service, method, path}
	m.activeConnections.WithLabelValues(labels...).Set(count)
}

// IncActiveConnections increments the active connections counter
func (m *Metrics) IncActiveConnections(service, method, path string) {
	labels := []string{service, method, path}
	m.activeConnections.WithLabelValues(labels...).Inc()
}

// DecActiveConnections decrements the active connections counter
func (m *Metrics) DecActiveConnections(service, method, path string) {
	labels := []string{service, method, path}
	m.activeConnections.WithLabelValues(labels...).Dec()
}
