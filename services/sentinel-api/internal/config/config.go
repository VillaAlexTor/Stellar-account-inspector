package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address              string
	DatabaseURL          string
	AllowedOrigins       []string
	AuthTokens           []string
	SessionSecret        string
	SessionTTL           time.Duration
	SecureCookie         bool
	RateLimitWindow      time.Duration
	RateLimitIP          int
	RateLimitAccount     int
	TrustedProxyCIDRs    []string
	LogLevel             slog.Level
	Environment          string
	Version              string
	AlertRetention       time.Duration
	OperationRetention   time.Duration
	RetentionInterval    time.Duration
	WebhookURL           string
	WebhookSecret        string
	TelegramBotToken     string
	TelegramChatID       string
	TelegramAPIURL       string
	SMTPHost             string
	SMTPPort             int
	SMTPUsername         string
	SMTPPassword         string
	SMTPFrom             string
	SMTPTo               []string
	SMTPTLS              string
	NotificationTimeout  time.Duration
	NotificationPoll     time.Duration
	NotificationAttempts int
	HorizonTestnetURL    string
	HTTPTimeout          time.Duration
	ReconnectMaxBackoff  time.Duration
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
	logLevel, err := logLevelEnv("SENTINEL_LOG_LEVEL", slog.LevelInfo)
	if err != nil {
		return Config{}, err
	}
	alertRetention, err := durationEnv("SENTINEL_ALERT_RETENTION", 90*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	operationRetention, err := durationEnv("SENTINEL_OPERATION_RETENTION", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	retentionInterval, err := durationEnv("SENTINEL_RETENTION_INTERVAL", time.Hour)
	if err != nil {
		return Config{}, err
	}
	notificationTimeout, err := durationEnv("SENTINEL_NOTIFICATION_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	notificationPoll, err := durationEnv("SENTINEL_NOTIFICATION_POLL", 2*time.Second)
	if err != nil {
		return Config{}, err
	}
	notificationAttempts, err := positiveIntEnv("SENTINEL_NOTIFICATION_MAX_ATTEMPTS", 5)
	if err != nil {
		return Config{}, err
	}
	for key, value := range map[string]time.Duration{
		"SENTINEL_HTTP_TIMEOUT":          httpTimeout,
		"SENTINEL_RECONNECT_MAX_BACKOFF": maxBackoff,
		"SENTINEL_SESSION_TTL":           sessionTTL,
		"SENTINEL_RATE_LIMIT_WINDOW":     rateLimitWindow,
		"SENTINEL_ALERT_RETENTION":       alertRetention,
		"SENTINEL_OPERATION_RETENTION":   operationRetention,
		"SENTINEL_RETENTION_INTERVAL":    retentionInterval,
		"SENTINEL_NOTIFICATION_TIMEOUT":  notificationTimeout,
		"SENTINEL_NOTIFICATION_POLL":     notificationPoll,
	} {
		if value <= 0 {
			return Config{}, fmt.Errorf("%s debe ser una duración positiva", key)
		}
	}
	smtpPort, err := positiveIntEnv("SENTINEL_SMTP_PORT", 587)
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
	environment := envOr("SENTINEL_ENVIRONMENT", "development")
	horizonTestnetURL := envOr("HORIZON_TESTNET_URL", "https://horizon-testnet.stellar.org")
	for key, value := range map[string]string{
		"HORIZON_TESTNET_URL": horizonTestnetURL,
	} {
		parsed, err := url.ParseRequestURI(value)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return Config{}, fmt.Errorf("%s debe ser una URL HTTP(S) válida", key)
		}
		if environment == "production" && parsed.Scheme != "https" {
			return Config{}, fmt.Errorf("%s debe usar HTTPS en producción", key)
		}
	}
	if environment == "production" {
		if len(authTokens) == 0 {
			return Config{}, fmt.Errorf("SENTINEL_AUTH_TOKENS es obligatorio en producción")
		}
		if !secureCookie {
			return Config{}, fmt.Errorf("SENTINEL_COOKIE_SECURE debe ser true en producción")
		}
		for _, origin := range origins {
			if origin == "*" {
				return Config{}, fmt.Errorf("SENTINEL_ALLOWED_ORIGINS no puede contener * en producción")
			}
		}
	}
	webhookURL := strings.TrimSpace(os.Getenv("SENTINEL_WEBHOOK_URL"))
	if webhookURL != "" {
		parsed, err := url.ParseRequestURI(webhookURL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return Config{}, fmt.Errorf("SENTINEL_WEBHOOK_URL debe ser una URL HTTP(S) válida")
		}
		if environment == "production" && parsed.Scheme != "https" {
			return Config{}, fmt.Errorf("SENTINEL_WEBHOOK_URL debe usar HTTPS en producción")
		}
	}
	telegramToken := strings.TrimSpace(os.Getenv("SENTINEL_TELEGRAM_BOT_TOKEN"))
	telegramChatID := strings.TrimSpace(os.Getenv("SENTINEL_TELEGRAM_CHAT_ID"))
	telegramAPIURL := envOr("SENTINEL_TELEGRAM_API_URL", "https://api.telegram.org")
	if (telegramToken == "") != (telegramChatID == "") {
		return Config{}, fmt.Errorf("SENTINEL_TELEGRAM_BOT_TOKEN y SENTINEL_TELEGRAM_CHAT_ID deben configurarse juntos")
	}
	if telegramToken != "" {
		parsed, err := url.ParseRequestURI(telegramAPIURL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return Config{}, fmt.Errorf("SENTINEL_TELEGRAM_API_URL debe ser una URL HTTP(S) válida")
		}
		if environment == "production" && parsed.Scheme != "https" {
			return Config{}, fmt.Errorf("SENTINEL_TELEGRAM_API_URL debe usar HTTPS en producción")
		}
	}
	smtpHost := strings.TrimSpace(os.Getenv("SENTINEL_SMTP_HOST"))
	smtpFrom := strings.TrimSpace(os.Getenv("SENTINEL_SMTP_FROM"))
	smtpTo := splitCSV(os.Getenv("SENTINEL_SMTP_TO"))
	if smtpHost != "" && (smtpFrom == "" || len(smtpTo) == 0) {
		return Config{}, fmt.Errorf("SENTINEL_SMTP_FROM y SENTINEL_SMTP_TO son obligatorios al configurar SMTP")
	}
	if smtpHost == "" && (smtpFrom != "" || len(smtpTo) > 0) {
		return Config{}, fmt.Errorf("SENTINEL_SMTP_HOST es obligatorio al configurar correo")
	}
	smtpTLS := strings.ToLower(envOr("SENTINEL_SMTP_TLS", "starttls"))
	if smtpTLS != "starttls" && smtpTLS != "implicit" && smtpTLS != "none" {
		return Config{}, fmt.Errorf("SENTINEL_SMTP_TLS debe ser starttls, implicit o none")
	}
	if environment == "production" && smtpHost != "" && smtpTLS == "none" {
		return Config{}, fmt.Errorf("SENTINEL_SMTP_TLS no puede ser none en producción")
	}
	trustedProxyCIDRs := splitCSV(os.Getenv("SENTINEL_TRUSTED_PROXY_CIDRS"))
	for _, value := range trustedProxyCIDRs {
		if _, _, err := net.ParseCIDR(value); err != nil {
			return Config{}, fmt.Errorf("SENTINEL_TRUSTED_PROXY_CIDRS contiene %q inválido: %w", value, err)
		}
	}
	return Config{
		Address:              envOr("SENTINEL_ADDRESS", ":8080"),
		DatabaseURL:          databaseURL,
		AllowedOrigins:       origins,
		AuthTokens:           authTokens,
		SessionSecret:        sessionSecret,
		SessionTTL:           sessionTTL,
		SecureCookie:         secureCookie,
		RateLimitWindow:      rateLimitWindow,
		RateLimitIP:          rateLimitIP,
		RateLimitAccount:     rateLimitAccount,
		TrustedProxyCIDRs:    trustedProxyCIDRs,
		LogLevel:             logLevel,
		Environment:          environment,
		Version:              envOr("SENTINEL_VERSION", "dev"),
		AlertRetention:       alertRetention,
		OperationRetention:   operationRetention,
		RetentionInterval:    retentionInterval,
		WebhookURL:           webhookURL,
		WebhookSecret:        strings.TrimSpace(os.Getenv("SENTINEL_WEBHOOK_SECRET")),
		TelegramBotToken:     telegramToken,
		TelegramChatID:       telegramChatID,
		TelegramAPIURL:       telegramAPIURL,
		SMTPHost:             smtpHost,
		SMTPPort:             smtpPort,
		SMTPUsername:         strings.TrimSpace(os.Getenv("SENTINEL_SMTP_USERNAME")),
		SMTPPassword:         os.Getenv("SENTINEL_SMTP_PASSWORD"),
		SMTPFrom:             smtpFrom,
		SMTPTo:               smtpTo,
		SMTPTLS:              smtpTLS,
		NotificationTimeout:  notificationTimeout,
		NotificationPoll:     notificationPoll,
		NotificationAttempts: notificationAttempts,
		HorizonTestnetURL:    horizonTestnetURL,
		HTTPTimeout:          httpTimeout,
		ReconnectMaxBackoff:  maxBackoff,
	}, nil
}

func logLevelEnv(key string, fallback slog.Level) (slog.Level, error) {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback, nil
	}
	levels := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	level, exists := levels[value]
	if !exists {
		return 0, fmt.Errorf("%s debe ser debug, info, warn o error", key)
	}
	return level, nil
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
