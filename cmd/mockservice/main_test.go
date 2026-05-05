package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMuxForKindReturnsFintechMocks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind string
		path string
		want string
	}{
		{name: "user", kind: "user", path: "/user", want: "kyc_status"},
		{name: "payment", kind: "payment", path: "/payment-orders", want: "payment_orders"},
		{name: "account", kind: "account", path: "/account", want: "available_balance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mux, err := muxForKind(tt.kind)
			if err != nil {
				t.Fatalf("muxForKind error = %v", err)
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q", got)
			}
			if !json.Valid(rec.Body.Bytes()) {
				t.Fatalf("JSONが不正です: %s", rec.Body.String())
			}
			if body := rec.Body.String(); !strings.Contains(body, tt.want) {
				t.Fatalf("body = %s, want field %q", body, tt.want)
			}
		})
	}
}

func TestMuxForKindRejectsUnknownKind(t *testing.T) {
	t.Parallel()

	if _, err := muxForKind("unknown"); err == nil {
		t.Fatal("unknown kind should return error")
	}
}

func TestMockServiceEnvFallback(t *testing.T) {
	t.Parallel()

	if got := env("MISSING_MOCK_ENV", "fallback"); got != "fallback" {
		t.Fatalf("env = %q, want fallback", got)
	}
}

func TestMockServiceEnvOverride(t *testing.T) {
	t.Setenv("MOCK_ENV", "override")

	if got := env("MOCK_ENV", "fallback"); got != "override" {
		t.Fatalf("env = %q, want override", got)
	}
}
