package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr         string
	DatabaseURL      string
	JWTSecret        string
	TelegramToken    string
	HHBaseURL        string
	PollInterval     time.Duration
	VacanciesPerPage int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		TelegramToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		HHBaseURL:        getEnv("HH_BASE_URL", "https://api.hh.ru"),
		VacanciesPerPage: 20,
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}

	interval := getEnv("POLL_INTERVAL", "5m")
	d, err := time.ParseDuration(interval)
	if err != nil {
		return cfg, fmt.Errorf("invalid POLL_INTERVAL: %w", err)
	}
	cfg.PollInterval = d

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
