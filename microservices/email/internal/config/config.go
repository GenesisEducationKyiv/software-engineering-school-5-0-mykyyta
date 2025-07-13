package config

import (
	"github.com/joho/godotenv"
	"os"
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
		Port:        getEnv("PORT", "8081"),
		SendGridKey: mustGet("SENDGRID_API_KEY"),
		EmailFrom:   mustGet("EMAIL_FROM"),
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
