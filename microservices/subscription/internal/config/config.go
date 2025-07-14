package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	GinMode         string
	Port            string
	DBUrl           string
	JWTSecret       string
	BaseURL         string
	EmailAPIBaseURL string
	UseGRPC         bool
	WeatherGRPCAddr string
	WeatherHTTPAddr string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		GinMode:         getEnv("GIN_MODE", "debug"),
		Port:            getEnv("PORT", "8080"),
		DBUrl:           getEnv("DB_URL", "postgres://postgres:postgres@db:5432/weatherdb?sslmode=disable"),
		BaseURL:         strings.TrimRight(getEnv("BASE_URL", "http://localhost:8080"), "/"),
		JWTSecret:       mustGet("JWT_SECRET"),
		EmailAPIBaseURL: mustGet("EMAIL_API_BASE_URL"),
		UseGRPC:         getBoolEnv("USE_GRPC", true),
		WeatherGRPCAddr: getEnv("WEATHER_GRPC_ADDR", "weather:50051"),
		WeatherHTTPAddr: getEnv("WEATHER_HTTP_ADDR", "http://weather:8082"),
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
