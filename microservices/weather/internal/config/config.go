package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	GRPCPort         string
	WeatherAPIKey    string
	TomorrowioAPIKey string
	Cache            CacheConfig
	BenchmarkMode    bool
}

type CacheConfig struct {
	Enabled       bool
	RedisURL      string
	WeatherApiTTL time.Duration
	TomorrowIoTTL time.Duration
	NotFoundTTL   time.Duration
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:             getEnv("PORT", "8082"),
		GRPCPort:         getEnv("GRPC_PORT", "50051"),
		WeatherAPIKey:    mustGet("WEATHER_API_KEY"),
		TomorrowioAPIKey: mustGet("TOMORROWIO_API_KEY"),
		Cache:            loadCacheConfig(),
		BenchmarkMode:    getBoolEnv("BENCHMARK_MODE", false),
	}
}

func loadCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:       getBoolEnv("CACHE_ENABLED", true),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
		WeatherApiTTL: getDurationEnv("CACHE_TTL_WEATHERAPI", 15*time.Minute),
		TomorrowIoTTL: getDurationEnv("CACHE_TTL_TOMORROWIO", 2*time.Minute),
		NotFoundTTL:   getDurationEnv("CACHE_TTL_NOTFOUND", 1*time.Minute),
	}
}

func mustGet(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("Missing required environment variable: " + key)
	}
	return val
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getBoolEnv(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	result, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return result
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	dur, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return dur
}
