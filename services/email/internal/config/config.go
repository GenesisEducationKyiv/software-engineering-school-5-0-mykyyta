package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	SendGridKey string
	EmailFrom   string
	BaseURL     string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		SendGridKey: mustGet("SENDGRID_API_KEY"),
		EmailFrom:   mustGet("EMAIL_FROM"),
		BaseURL:     strings.TrimRight(getEnv("BASE_URL", "http://localhost:8080"), "/"),
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
