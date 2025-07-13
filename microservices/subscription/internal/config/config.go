package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	GinMode           string
	Port              string
	DBUrl             string
	JWTSecret         string
	BaseURL           string
	EmailAPIBaseURL   string
	EmailFrom         string
	WeatherAPIBaseURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		GinMode:           getEnv("GIN_MODE", "debug"),
		Port:              getEnv("PORT", "8080"),
		DBUrl:             getEnv("DB_URL", "postgres://postgres:postgres@db:5432/weatherdb?sslmode=disable"),
		BaseURL:           strings.TrimRight(getEnv("BASE_URL", "http://localhost:8080"), "/"),
		JWTSecret:         mustGet("JWT_SECRET"),
		EmailAPIBaseURL:   mustGet("EMAIL_API_BASE_URL"),
		EmailFrom:         mustGet("EMAIL_FROM"),
		WeatherAPIBaseURL: mustGet("WEATHER_API_BASE_URL"),
	}
}

// --- utils ---

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
