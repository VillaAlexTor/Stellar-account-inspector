package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address             string
	DatabaseURL         string
	AllowedOrigins      []string
	AuthTokens          []string
	SessionSecret       string
	SessionTTL          time.Duration
	SecureCookie        bool
	RateLimitWindow     time.Duration
	RateLimitIP         int
	RateLimitAccount    int
	TrustedProxyCIDRs   []string
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
	sessionTTL, err := durationEnv("SENTINEL_SESSION_TTL", 12*time.Hour)
	if err != nil {
		return Config{}, err
	}
	rateLimitWindow, err := durationEnv("SENTINEL_RATE_LIMIT_WINDOW", time.Minute)
	if err != nil {
		return Config{}, err
	}
	rateLimitIP, err := positiveIntEnv("SENTINEL_RATE_LIMIT_IP", 120)
	if err != nil {
		return Config{}, err
	}
	rateLimitAccount, err := positiveIntEnv("SENTINEL_RATE_LIMIT_ACCOUNT", 30)
	if err != nil {
		return Config{}, err
	}
	secureCookie, err := boolEnv("SENTINEL_COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL es obligatorio")
	}

	origins := splitCSV(envOr("SENTINEL_ALLOWED_ORIGINS", "http://localhost:3000"))
	authTokens := splitCSV(os.Getenv("SENTINEL_AUTH_TOKENS"))
	sessionSecret := strings.TrimSpace(os.Getenv("SENTINEL_SESSION_SECRET"))
	if len(authTokens) > 0 && len(sessionSecret) < 32 {
		return Config{}, fmt.Errorf("SENTINEL_SESSION_SECRET debe tener al menos 32 caracteres cuando la autenticación está activa")
	}
	trustedProxyCIDRs := splitCSV(os.Getenv("SENTINEL_TRUSTED_PROXY_CIDRS"))
	for _, value := range trustedProxyCIDRs {
		if _, _, err := net.ParseCIDR(value); err != nil {
			return Config{}, fmt.Errorf("SENTINEL_TRUSTED_PROXY_CIDRS contiene %q inválido: %w", value, err)
		}
	}
	return Config{
		Address:             envOr("SENTINEL_ADDRESS", ":8080"),
		DatabaseURL:         databaseURL,
		AllowedOrigins:      origins,
		AuthTokens:          authTokens,
		SessionSecret:       sessionSecret,
		SessionTTL:          sessionTTL,
		SecureCookie:        secureCookie,
		RateLimitWindow:     rateLimitWindow,
		RateLimitIP:         rateLimitIP,
		RateLimitAccount:    rateLimitAccount,
		TrustedProxyCIDRs:   trustedProxyCIDRs,
		HorizonTestnetURL:   envOr("HORIZON_TESTNET_URL", "https://horizon-testnet.stellar.org"),
		HorizonMainnetURL:   envOr("HORIZON_MAINNET_URL", "https://horizon.stellar.org"),
		HTTPTimeout:         httpTimeout,
		ReconnectMaxBackoff: maxBackoff,
	}, nil
}

func positiveIntEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s debe ser un entero positivo", key)
	}
	return parsed, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s inválido: %w", key, err)
	}
	return parsed, nil
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
