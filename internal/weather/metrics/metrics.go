package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_hits_total",
			Help: "Total number of cache hits per provider",
		},
		[]string{"provider"},
	)

	CacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_misses_total",
			Help: "Total number of cache misses per provider",
		},
		[]string{"provider"},
	)
	CacheResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_cache_result_total",
			Help: "Overall cache result for a GetWeather request",
		},
		[]string{"status"},
	)
)

func Register() {
	prometheus.MustRegister(CacheHits, CacheMisses, CacheResult)
}
