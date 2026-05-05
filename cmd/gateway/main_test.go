package main

import (
	"log/slog"
	"testing"
	"time"
)

func TestConfigFromEnvUsesDefaults(t *testing.T) {
	t.Parallel()

	addr, cfg := configFromEnv(slog.Default())

	if addr != ":8080" {
		t.Fatalf("addr = %q, want %q", addr, ":8080")
	}
	if cfg.UserURL != "http://user-service:8081/user" {
		t.Fatalf("UserURL = %q", cfg.UserURL)
	}
	if cfg.PaymentURL != "http://payment-service:8082/payment-orders" {
		t.Fatalf("PaymentURL = %q", cfg.PaymentURL)
	}
	if cfg.AccountURL != "http://account-service:8083/account" {
		t.Fatalf("AccountURL = %q", cfg.AccountURL)
	}
	if cfg.Timeout != 800*time.Millisecond {
		t.Fatalf("Timeout = %v", cfg.Timeout)
	}
	if cfg.CacheTTL != 5*time.Second {
		t.Fatalf("CacheTTL = %v", cfg.CacheTTL)
	}
	if cfg.RateLimit.Burst != 100 || cfg.RateLimit.Refill != 100 || cfg.RateLimit.Every != time.Second {
		t.Fatalf("RateLimit = %+v", cfg.RateLimit)
	}
}

func TestConfigFromEnvUsesOverrides(t *testing.T) {
	t.Setenv("ADDR", ":9090")
	t.Setenv("USER_SERVICE_URL", "http://127.0.0.1:18081/user")
	t.Setenv("PAYMENT_SERVICE_URL", "http://127.0.0.1:18082/payment-orders")
	t.Setenv("ACCOUNT_SERVICE_URL", "http://127.0.0.1:18083/account")
	t.Setenv("DOWNSTREAM_TIMEOUT", "250ms")
	t.Setenv("CACHE_TTL", "2s")
	t.Setenv("RATE_LIMIT_BURST", "7")
	t.Setenv("RATE_LIMIT_REFILL", "3")
	t.Setenv("RATE_LIMIT_EVERY", "500ms")

	addr, cfg := configFromEnv(slog.Default())

	if addr != ":9090" {
		t.Fatalf("addr = %q, want %q", addr, ":9090")
	}
	if cfg.UserURL != "http://127.0.0.1:18081/user" {
		t.Fatalf("UserURL = %q", cfg.UserURL)
	}
	if cfg.PaymentURL != "http://127.0.0.1:18082/payment-orders" {
		t.Fatalf("PaymentURL = %q", cfg.PaymentURL)
	}
	if cfg.AccountURL != "http://127.0.0.1:18083/account" {
		t.Fatalf("AccountURL = %q", cfg.AccountURL)
	}
	if cfg.Timeout != 250*time.Millisecond {
		t.Fatalf("Timeout = %v", cfg.Timeout)
	}
	if cfg.CacheTTL != 2*time.Second {
		t.Fatalf("CacheTTL = %v", cfg.CacheTTL)
	}
	if cfg.RateLimit.Burst != 7 || cfg.RateLimit.Refill != 3 || cfg.RateLimit.Every != 500*time.Millisecond {
		t.Fatalf("RateLimit = %+v", cfg.RateLimit)
	}
}

func TestEnvHelpersFallbackInvalidValues(t *testing.T) {
	t.Setenv("INVALID_INT", "invalid")
	t.Setenv("INVALID_DURATION", "invalid")

	if got := intEnv("INVALID_INT", 42); got != 42 {
		t.Fatalf("intEnv = %d, want 42", got)
	}
	if got := durationEnv("INVALID_DURATION", 3*time.Second); got != 3*time.Second {
		t.Fatalf("durationEnv = %v, want 3s", got)
	}
}
