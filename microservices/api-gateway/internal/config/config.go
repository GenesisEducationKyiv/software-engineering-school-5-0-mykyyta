package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	Port                    string        `env:"PORT"`
	SubscriptionServiceAddr string        `env:"SUBSCRIPTION_SERVICE_ADDR"`
	RequestTimeout          time.Duration `env:"REQUEST_TIMEOUT"`
	ReadTimeout             time.Duration `env:"READ_TIMEOUT"`
	WriteTimeout            time.Duration `env:"WRITE_TIMEOUT"`
}

func LoadConfig() *Config {
	return &Config{
		Port:                    getEnv("PORT", "8080"),
		SubscriptionServiceAddr: getEnv("SUBSCRIPTION_SERVICE_ADDR", "http://localhost:8083"),
		RequestTimeout:          parseDuration(getEnv("REQUEST_TIMEOUT", "30s")),
		ReadTimeout:             parseDuration(getEnv("READ_TIMEOUT", "15s")),
		WriteTimeout:            parseDuration(getEnv("WRITE_TIMEOUT", "15s")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("Invalid duration %s, using default", s)
		return 30 * time.Second
	}
	return d
}
