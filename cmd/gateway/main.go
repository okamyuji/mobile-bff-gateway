package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/okamyuji/mobile-bff-gateway/internal/gateway"
)

// main はGatewayサーバーを起動します。
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	addr, cfg := configFromEnv(logger)

	server := &http.Server{
		Addr:              addr,
		Handler:           gateway.New(cfg),
		ReadHeaderTimeout: 3 * time.Second,
	}
	logger.Info("Gatewayを起動します", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Gatewayを起動できませんでした", "error", err)
		os.Exit(1)
	}
}

func configFromEnv(logger *slog.Logger) (string, gateway.Config) {
	cfg := gateway.Config{
		UserURL:    env("USER_SERVICE_URL", "http://user-service:8081/user"),
		PaymentURL: env("PAYMENT_SERVICE_URL", "http://payment-service:8082/payment-orders"),
		AccountURL: env("ACCOUNT_SERVICE_URL", "http://account-service:8083/account"),
		Timeout:    durationEnv("DOWNSTREAM_TIMEOUT", 800*time.Millisecond),
		CacheTTL:   durationEnv("CACHE_TTL", 5*time.Second),
		RateLimit: gateway.RateLimitConfig{
			Burst:  intEnv("RATE_LIMIT_BURST", 100),
			Refill: intEnv("RATE_LIMIT_REFILL", 100),
			Every:  durationEnv("RATE_LIMIT_EVERY", time.Second),
		},
		Logger: logger,
	}
	return env("ADDR", ":8080"), cfg
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func intEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
