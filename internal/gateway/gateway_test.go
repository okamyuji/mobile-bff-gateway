package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/okamyuji/mobile-bff-gateway/internal/auth"
)

func TestMobileHomeAggregatesAndShrinksJSON(t *testing.T) {
	t.Parallel()

	client := mockClient(map[string]string{
		"user":    `{"id":"user-123","name":"佐藤","email":"sato@example.com","internal_note":"不要"}`,
		"payment": `{"payment_orders":[{"id":"pay-1","status":"authorized","amount":1200,"risk_score":12}]}`,
		"account": `{"account":{"id":"acct-1","available_balance":120000,"currency":"JPY","ledger_version":"internal"}}`,
	}, nil)

	handler := New(Config{
		UserURL:    "https://user",
		PaymentURL: "https://payment",
		AccountURL: "https://account",
		Timeout:    time.Second,
		CacheTTL:   time.Minute,
		RateLimit:  RateLimitConfig{Burst: 10, Refill: 10, Every: time.Second},
		Client:     client,
	})

	req := httptest.NewRequest(http.MethodGet, "/mobile/home", nil)
	req.Header.Set("Authorization", "Bearer "+auth.MakeTestToken(t, time.Now().Add(time.Hour)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() > 512 {
		t.Fatalf("レスポンスが大きすぎます: %d bytes", rec.Body.Len())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSONを解析できませんでした: %v", err)
	}
	if _, ok := body["email"]; ok {
		t.Fatal("モバイルJSONに不要なemailが含まれています")
	}
	if _, ok := body["internal_note"]; ok {
		t.Fatal("モバイルJSONに内部項目が含まれています")
	}
	if _, ok := body["ledger_version"]; ok {
		t.Fatal("モバイルJSONに内部台帳項目が含まれています")
	}
}

func TestMobileHomeUsesCache(t *testing.T) {
	t.Parallel()

	var userCalls int64
	client := mockClient(map[string]string{
		"user":    `{"id":"user-123","name":"佐藤"}`,
		"payment": `{"payment_orders":[]}`,
		"account": `{"account":{"available_balance":120000,"currency":"JPY"}}`,
	}, map[string]*int64{"user": &userCalls})

	handler := New(Config{
		UserURL:    "https://user",
		PaymentURL: "https://payment",
		AccountURL: "https://account",
		Timeout:    time.Second,
		CacheTTL:   time.Minute,
		RateLimit:  RateLimitConfig{Burst: 10, Refill: 10, Every: time.Second},
		Client:     client,
	})

	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/mobile/home", nil)
		req.Header.Set("Authorization", "Bearer "+auth.MakeTestToken(t, time.Now().Add(time.Hour)))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}
	}

	if got := atomic.LoadInt64(&userCalls); got != 1 {
		t.Fatalf("user-service calls = %d, want 1", got)
	}
}

func TestMobileHomeRejectsMissingBearerToken(t *testing.T) {
	t.Parallel()

	handler := New(Config{
		UserURL:    "https://user",
		PaymentURL: "https://payment",
		AccountURL: "https://account",
		RateLimit:  RateLimitConfig{Burst: 10, Refill: 10, Every: time.Second},
		Client:     mockClient(nil, nil),
	})

	req := httptest.NewRequest(http.MethodGet, "/mobile/home", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestServeHTTPHealthz(t *testing.T) {
	t.Parallel()

	handler := New(Config{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestServeHTTPUnknownPath(t *testing.T) {
	t.Parallel()

	handler := New(Config{})
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestMobileHomeRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	handler := New(Config{})
	req := httptest.NewRequest(http.MethodPost, "/mobile/home", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestMobileHomeAppliesRateLimit(t *testing.T) {
	t.Parallel()

	handler := New(Config{
		UserURL:    "https://user",
		PaymentURL: "https://payment",
		AccountURL: "https://account",
		RateLimit:  RateLimitConfig{Burst: 1, Refill: 0, Every: time.Hour},
		Client: mockClient(map[string]string{
			"user":    `{"id":"user-123","name":"佐藤"}`,
			"payment": `{"payment_orders":[]}`,
			"account": `{"account":{"available_balance":120000,"currency":"JPY"}}`,
		}, nil),
	})

	for index := range 2 {
		req := httptest.NewRequest(http.MethodGet, "/mobile/home", nil)
		req.Header.Set("Authorization", "Bearer "+auth.MakeTestToken(t, time.Now().Add(time.Hour)))
		req.RemoteAddr = "192.0.2.10:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if index == 0 && rec.Code != http.StatusOK {
			t.Fatalf("first status = %d", rec.Code)
		}
		if index == 1 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("second status = %d, want %d", rec.Code, http.StatusTooManyRequests)
		}
	}
}

func TestMobileHomeReturnsBadGatewayWhenDownstreamFails(t *testing.T) {
	t.Parallel()

	handler := New(Config{
		UserURL:    "https://user",
		PaymentURL: "https://payment",
		AccountURL: "https://account",
		Timeout:    time.Second,
		RateLimit:  RateLimitConfig{Burst: 10, Refill: 10, Every: time.Second},
		Client:     errorClient(errors.New("接続できません")),
	})

	req := httptest.NewRequest(http.MethodGet, "/mobile/home", nil)
	req.Header.Set("Authorization", "Bearer "+auth.MakeTestToken(t, time.Now().Add(time.Hour)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func mockClient(bodies map[string]string, calls map[string]*int64) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if counter := calls[req.URL.Host]; counter != nil {
			atomic.AddInt64(counter, 1)
		}
		body := bodies[req.URL.Host]
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Request:    req,
		}, nil
	})}
}

func errorClient(err error) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, err
	})}
}
