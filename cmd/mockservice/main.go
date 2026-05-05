package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// main は指定された種類のモックHTTPサービスを起動します。
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	kind := env("SERVICE_KIND", "user")
	addr := env("ADDR", ":8081")

	mux, err := muxForKind(kind)
	if err != nil {
		logger.Error("未知のモックサービス種別です", "kind", kind)
		os.Exit(1)
	}
	mux.HandleFunc("/healthz", writeStatic(`{"status":"ok"}`))

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	logger.Info("モックサービスを起動します", "kind", kind, "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("モックサービスを起動できませんでした", "error", err)
		os.Exit(1)
	}
}

func muxForKind(kind string) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	switch kind {
	case "user":
		mux.HandleFunc("/user", writeStatic(`{"id":"user-123","name":"佐藤","email":"sato@example.com","kyc_status":"verified","internal_note":"gatewayで削除します"}`))
	case "payment":
		mux.HandleFunc("/payment-orders", writeStatic(`{"payment_orders":[{"id":"pay-1","status":"authorized","amount":1200,"risk_score":12},{"id":"pay-2","status":"settled","amount":980,"risk_score":4}]}`))
	case "account":
		mux.HandleFunc("/account", writeStatic(`{"account":{"id":"acct-1","available_balance":120000,"currency":"JPY","ledger_version":"internal-2026-05-05"}}`))
	default:
		return nil, fmt.Errorf("unknown service kind: %s", kind)
	}
	return mux, nil
}

func writeStatic(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
