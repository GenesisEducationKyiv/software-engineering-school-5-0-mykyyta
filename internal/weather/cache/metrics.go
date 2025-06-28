package cache

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	Access = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_access_total",
			Help: "Total cache access attempts by provider and status",
		},
		[]string{"provider", "status"}, // status = "hit" | "miss"
	)

	Result = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_result_total",
			Help: "Overall cache result for full GetWeather request",
		},
		[]string{"status"}, // status = "hit" | "miss"
	)
)

func Register() {
	prometheus.MustRegister(Access, Result)
}

func RecordProvider(provider, status string) {
	Access.WithLabelValues(provider, status).Inc()
}

func RecordTotal(status string) {
	Result.WithLabelValues(status).Inc()
}
