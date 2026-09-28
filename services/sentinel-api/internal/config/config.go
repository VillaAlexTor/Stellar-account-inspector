package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address             string
	DatabaseURL         string
	AllowedOrigins      []string
	HorizonTestnetURL   string
	HorizonMainnetURL   string
	HTTPTimeout         time.Duration
	ReconnectMaxBackoff time.Duration
}

func Load() (Config, error) {
	httpTimeout, err := durationEnv("SENTINEL_HTTP_TIMEOUT", 45*time.Second)
	if err != nil {
		return Config{}, err
	}
	maxBackoff, err := durationEnv("SENTINEL_RECONNECT_MAX_BACKOFF", 30*time.Second)
	if err != nil {
		return Config{}, err
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL es obligatorio")
	}

	origins := splitCSV(envOr("SENTINEL_ALLOWED_ORIGINS", "http://localhost:3000"))
	return Config{
		Address:             envOr("SENTINEL_ADDRESS", ":8080"),
		DatabaseURL:         databaseURL,
		AllowedOrigins:      origins,
		HorizonTestnetURL:   envOr("HORIZON_TESTNET_URL", "https://horizon-testnet.stellar.org"),
		HorizonMainnetURL:   envOr("HORIZON_MAINNET_URL", "https://horizon.stellar.org"),
		HTTPTimeout:         httpTimeout,
		ReconnectMaxBackoff: maxBackoff,
	}, nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s inválido: %w", key, err)
	}
	return parsed, nil
}
