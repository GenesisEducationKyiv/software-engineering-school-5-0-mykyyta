package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	SendGridKey string
	EmailFrom   string
	BaseURL     string
	RedisURL    string
	GmailPass   string
	GmailAddr   string
	RabbitMQURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8081"),
		SendGridKey: mustGet("SENDGRID_API_KEY"),
		EmailFrom:   mustGet("EMAIL_FROM"),
		RedisURL:    getEnv("REDIS_URL", "redis://redis:6379/1"),
		GmailPass:   mustGet("GMAIL_PASSWORD"),
		GmailAddr:   mustGet("GMAIL_ADDRESS"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
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
