package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port                    string        `env:"PORT"`
	SubscriptionServiceAddr string        `env:"SUBSCRIPTION_SERVICE_ADDR"`
	RequestTimeout          time.Duration `env:"REQUEST_TIMEOUT"`
	ReadTimeout             time.Duration `env:"READ_TIMEOUT"`
	WriteTimeout            time.Duration `env:"WRITE_TIMEOUT"`

	Service string  `yaml:"service"`
	Version string  `yaml:"version"`
	Routes  []Route `yaml:"routes"`
}

type Route struct {
	Path         string `yaml:"path"`
	Method       string `yaml:"method"`
	Handler      string `yaml:"handler"`
	AuthRequired bool   `yaml:"auth_required"`
	RateLimit    int    `yaml:"rate_limit,omitempty"`
	Cache        bool   `yaml:"cache,omitempty"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                    getEnv("PORT", "8080"),
		SubscriptionServiceAddr: getEnv("SUBSCRIPTION_SERVICE_ADDR", "http://localhost:8083"),
		RequestTimeout:          parseDuration(getEnv("REQUEST_TIMEOUT", "30s")),
		ReadTimeout:             parseDuration(getEnv("READ_TIMEOUT", "15s")),
		WriteTimeout:            parseDuration(getEnv("WRITE_TIMEOUT", "15s")),
	}

	yamlFile := getEnv("ROUTES_CONFIG_PATH", "config.yaml")
	if err := loadRoutesFromYAML(cfg, yamlFile); err != nil {
		return nil, err
	}

	return cfg, nil
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

func loadRoutesFromYAML(cfg *Config, path string) error {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open routes config file: %v", err)
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Failed to close routes config file: %v", err)
		}
	}()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(cfg); err != nil {
		log.Printf("Failed to decode YAML config: %v", err)
		return err
	}

	return nil
}
