package cache

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	access *prometheus.CounterVec
	result *prometheus.CounterVec
	once   sync.Once
}

func NewMetrics() *Metrics {
	return &Metrics{
		access: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_cache_access_total",
				Help: "Total cache access attempts by provider and status",
			},
			[]string{"provider", "status"},
		),
		result: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_cache_result_total",
				Help: "Overall cache result for full GetWeather request",
			},
			[]string{"status"},
		),
	}
}

func (m *Metrics) Register() {
	m.once.Do(func() {
		prometheus.MustRegister(m.access, m.result)
	})
}

func (m *Metrics) RecordProviderHit(provider string) {
	m.access.WithLabelValues(provider, "hit").Inc()
}

func (m *Metrics) RecordProviderMiss(provider string) {
	m.access.WithLabelValues(provider, "miss").Inc()
}

func (m *Metrics) RecordTotalHit() {
	m.result.WithLabelValues("hit").Inc()
}

func (m *Metrics) RecordTotalMiss() {
	m.result.WithLabelValues("miss").Inc()
}
