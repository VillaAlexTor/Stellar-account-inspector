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

func setRequiredTestEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("SENTINEL_AUTH_TOKENS", "")
	t.Setenv("SENTINEL_SESSION_SECRET", "")
	t.Setenv("SENTINEL_RATE_LIMIT_IP", "")
	t.Setenv("SENTINEL_RATE_LIMIT_ACCOUNT", "")
	t.Setenv("SENTINEL_TRUSTED_PROXY_CIDRS", "")
}
