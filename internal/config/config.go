package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	GinMode          string
	Port             string
	DBType           string
	DBUrl            string
	JWTSecret        string
	SendGridKey      string
	EmailFrom        string
	WeatherAPIKey    string
	TomorrowioAPIKey string
	BaseURL          string
	Cache            CacheConfig
}

type CacheConfig struct {
	Enabled        bool
	RedisURL       string
	DefaultTTL     time.Duration
	NotFoundTTL    time.Duration
	OpenWeatherTTL time.Duration
	WeatherApiTTL  time.Duration
	TomorrowIoTTL  time.Duration
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		GinMode:          getEnv("GIN_MODE", "debug"),
		Port:             getEnv("PORT", "8080"),
		DBType:           getEnv("DB_TYPE", "postgres"),
		DBUrl:            getEnv("DB_URL", "postgres://postgres:postgres@db:5432/weatherdb?sslmode=disable"),
		BaseURL:          strings.TrimRight(getEnv("BASE_URL", "http://localhost:8080"), "/"),
		JWTSecret:        getEnv("JWT_SECRET", "default_secret"),
		SendGridKey:      mustGet("SENDGRID_API_KEY"),
		EmailFrom:        mustGet("EMAIL_FROM"),
		WeatherAPIKey:    mustGet("WEATHER_API_KEY"),
		TomorrowioAPIKey: mustGet("TOMORROWIO_API_KEY"),
		Cache:            loadCacheConfig(),
	}
}

func loadCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:        getBoolEnv("CACHE_ENABLED", true),
		RedisURL:       getEnv("REDIS_URL", "redis://redis:6379/0"),
		DefaultTTL:     getDurationEnv("CACHE_TTL", 2*time.Minute),
		NotFoundTTL:    getDurationEnv("CACHE_TTL_NOTFOUND", 1*time.Minute),
		WeatherApiTTL:  getDurationEnv("CACHE_TTL_WEATHERAPI", 15*time.Minute),
		TomorrowIoTTL:  getDurationEnv("CACHE_TTL_TOMORROWIO", 2*time.Minute),
		OpenWeatherTTL: getDurationEnv("CACHE_TTL_OPENWEATHER", 10*time.Minute),
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
