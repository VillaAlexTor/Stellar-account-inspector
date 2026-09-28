package config

import (
	"strings"
	"testing"
)

func TestLoadSecurityConfiguration(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("SENTINEL_AUTH_TOKENS", "first-token, second-token")
	t.Setenv("SENTINEL_SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("SENTINEL_RATE_LIMIT_IP", "75")
	t.Setenv("SENTINEL_RATE_LIMIT_ACCOUNT", "12")
	t.Setenv("SENTINEL_TRUSTED_PROXY_CIDRS", "10.0.0.0/8,192.168.0.0/16")

	configuration, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(configuration.AuthTokens) != 2 || configuration.RateLimitIP != 75 || configuration.RateLimitAccount != 12 {
		t.Fatalf("unexpected security configuration: %#v", configuration)
	}
	if len(configuration.TrustedProxyCIDRs) != 2 {
		t.Fatalf("trusted proxy CIDRs = %#v", configuration.TrustedProxyCIDRs)
	}
}

func TestLoadRequiresStrongSessionSecretWhenAuthEnabled(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("SENTINEL_AUTH_TOKENS", "token")
	t.Setenv("SENTINEL_SESSION_SECRET", "short")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "32 caracteres") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadRejectsInvalidRateLimitAndProxy(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("SENTINEL_RATE_LIMIT_IP", "0")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a zero IP limit")
	}

	t.Setenv("SENTINEL_RATE_LIMIT_IP", "120")
	t.Setenv("SENTINEL_TRUSTED_PROXY_CIDRS", "not-a-cidr")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an invalid trusted proxy CIDR")
	}
}

func TestLoadNotificationAndRetentionConfiguration(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("SENTINEL_ALERT_RETENTION", "48h")
	t.Setenv("SENTINEL_NOTIFICATION_MAX_ATTEMPTS", "7")
	t.Setenv("SENTINEL_WEBHOOK_URL", "https://hooks.example.test/sentinel")
	t.Setenv("SENTINEL_TELEGRAM_BOT_TOKEN", "bot-token")
	t.Setenv("SENTINEL_TELEGRAM_CHAT_ID", "123")

	configuration, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.AlertRetention.Hours() != 48 || configuration.NotificationAttempts != 7 {
		t.Fatalf("unexpected retention/notification configuration: %#v", configuration)
	}
}

func TestLoadRejectsUnsafeProductionNotifications(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("SENTINEL_ENVIRONMENT", "production")
	t.Setenv("SENTINEL_AUTH_TOKENS", "production-token")
	t.Setenv("SENTINEL_SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("SENTINEL_COOKIE_SECURE", "true")
	t.Setenv("SENTINEL_WEBHOOK_URL", "http://hooks.example.test/sentinel")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "HTTPS") {
		t.Fatalf("Load() webhook error = %v", err)
	}

	t.Setenv("SENTINEL_WEBHOOK_URL", "")
	t.Setenv("SENTINEL_SMTP_HOST", "mail.example.test")
	t.Setenv("SENTINEL_SMTP_FROM", "sentinel@example.test")
	t.Setenv("SENTINEL_SMTP_TO", "ops@example.test")
	t.Setenv("SENTINEL_SMTP_TLS", "none")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "no puede ser none") {
		t.Fatalf("Load() SMTP error = %v", err)
	}
}

func TestLoadRejectsNonPositiveDurations(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("SENTINEL_RETENTION_INTERVAL", "0s")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "duración positiva") {
		t.Fatalf("Load() duration error = %v", err)
	}
}

func setRequiredTestEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("SENTINEL_AUTH_TOKENS", "")
	t.Setenv("SENTINEL_SESSION_SECRET", "")
	t.Setenv("SENTINEL_RATE_LIMIT_IP", "")
	t.Setenv("SENTINEL_RATE_LIMIT_ACCOUNT", "")
	t.Setenv("SENTINEL_TRUSTED_PROXY_CIDRS", "")
	t.Setenv("SENTINEL_ENVIRONMENT", "development")
	t.Setenv("SENTINEL_COOKIE_SECURE", "false")
	t.Setenv("SENTINEL_WEBHOOK_URL", "")
	t.Setenv("SENTINEL_TELEGRAM_BOT_TOKEN", "")
	t.Setenv("SENTINEL_TELEGRAM_CHAT_ID", "")
	t.Setenv("SENTINEL_SMTP_HOST", "")
	t.Setenv("SENTINEL_SMTP_FROM", "")
	t.Setenv("SENTINEL_SMTP_TO", "")
	t.Setenv("SENTINEL_SMTP_TLS", "")
}
