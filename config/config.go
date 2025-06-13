package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DBType        string
	DBUrl         string
	JWTSecret     string
	SendGridKey   string
	EmailFrom     string
	WeatherAPIKey string
	BaseURL       string
}

var C *Config

func LoadConfig() {
	_ = godotenv.Load()

	C = &Config{
		Port:          getEnv("PORT", "8080"),
		DBType:        getEnv("DB_TYPE", "postgres"),
		DBUrl:         getEnv("DB_URL", "host=your-host user=your-user password=your-password dbname=your-db port=5432 sslmode=require"),
		BaseURL:       strings.TrimRight(getEnv("BASE_URL", "http://localhost:8080"), "/"),
		JWTSecret:     getEnv("JWT_SECRET", "default_secret"),
		SendGridKey:   mustGet("SENDGRID_API_KEY"),
		EmailFrom:     mustGet("EMAIL_FROM"),
		WeatherAPIKey: mustGet("WEATHER_API_KEY"),
	}
}

func Reload() {
	C = nil
	LoadConfig()
}

func mustGet(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Missing required environment variable: %s", key)
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
