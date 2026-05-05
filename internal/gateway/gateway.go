package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/okamyuji/mobile-bff-gateway/internal/auth"
	"github.com/okamyuji/mobile-bff-gateway/internal/breaker"
	"github.com/okamyuji/mobile-bff-gateway/internal/mobile"
	"github.com/okamyuji/mobile-bff-gateway/internal/ratelimit"
)

// RateLimitConfig はGatewayのレート制限設定です。
type RateLimitConfig struct {
	Burst  int
	Refill int
	Every  time.Duration
}

// Config はGatewayサーバーの起動設定です。
type Config struct {
	UserURL    string
	PaymentURL string
	AccountURL string
	Timeout    time.Duration
	CacheTTL   time.Duration
	RateLimit  RateLimitConfig
	Logger     *slog.Logger
	Client     *http.Client
}

// Server はモバイルBFF用のHTTPハンドラーです。
type Server struct {
	userURL    string
	paymentURL string
	accountURL string
	timeout    time.Duration
	cacheTTL   time.Duration
	client     *http.Client
	limiter    *ratelimit.Limiter
	logger     *slog.Logger
	cache      responseCache
	breakers   map[string]*breaker.CircuitBreaker
}

type responseCache struct {
	mu      sync.Mutex
	expires time.Time
	body    []byte
}

type userResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type orderServiceResponse struct {
	PaymentOrders []mobile.PaymentOrderSummary `json:"payment_orders"`
}

type accountServiceResponse struct {
	Account struct {
		AvailableBalance int    `json:"available_balance"`
		Currency         string `json:"currency"`
	} `json:"account"`
}

// New はGatewayのHTTPハンドラーを作成します。
func New(cfg Config) *Server {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 800 * time.Millisecond
	}
	cacheTTL := cfg.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Second
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	rl := cfg.RateLimit
	if rl.Burst == 0 {
		rl.Burst = 100
	}
	if rl.Refill == 0 {
		rl.Refill = rl.Burst
	}
	if rl.Every == 0 {
		rl.Every = time.Second
	}

	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	return &Server{
		userURL:    cfg.UserURL,
		paymentURL: cfg.PaymentURL,
		accountURL: cfg.AccountURL,
		timeout:    timeout,
		cacheTTL:   cacheTTL,
		client:     client,
		limiter:    ratelimit.NewLimiter(rl.Burst, rl.Refill, rl.Every),
		logger:     logger,
		breakers: map[string]*breaker.CircuitBreaker{
			"user":    breaker.NewCircuitBreaker(3, 2*time.Second),
			"payment": breaker.NewCircuitBreaker(3, 2*time.Second),
			"account": breaker.NewCircuitBreaker(3, 2*time.Second),
		},
	}
}

// ServeHTTP はGatewayのHTTPリクエストを処理します。
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/healthz":
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case "/mobile/home":
		s.handleMobileHome(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleMobileHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := auth.ValidateBearer(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !s.limiter.Allow(clientKey(r)) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}
	if cached, ok := s.cachedHome(); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, max-age=5")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(cached)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	userCh := callAsync(ctx, func(ctx context.Context) (userResponse, error) {
		return fetchJSON[userResponse](ctx, s, "user", s.userURL)
	})
	paymentCh := callAsync(ctx, func(ctx context.Context) (orderServiceResponse, error) {
		return fetchJSON[orderServiceResponse](ctx, s, "payment", s.paymentURL)
	})
	accountCh := callAsync(ctx, func(ctx context.Context) (accountServiceResponse, error) {
		return fetchJSON[accountServiceResponse](ctx, s, "account", s.accountURL)
	})

	userResult := <-userCh
	paymentResult := <-paymentCh
	accountResult := <-accountCh
	if err := errors.Join(userResult.err, paymentResult.err, accountResult.err); err != nil {
		s.logger.Warn("下流サービスの呼び出しに失敗しました", "error", err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	body := mobile.HomeResponse{
		User:          mobile.UserSummary{ID: userResult.value.ID, Name: userResult.value.Name},
		PaymentOrders: paymentResult.value.PaymentOrders,
		Account: mobile.AccountSummary{
			AvailableBalance: accountResult.value.Account.AvailableBalance,
			Currency:         accountResult.value.Account.Currency,
		},
		Meta: mobile.ResponseMetadata{Cached: false},
	}
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	s.storeHome(data)
	s.logger.Info("モバイルホームを返しました", "subject", claims.Subject, "bytes", len(data))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, max-age=5")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

type asyncResult[T any] struct {
	value T
	err   error
}

func callAsync[T any](ctx context.Context, fn func(context.Context) (T, error)) <-chan asyncResult[T] {
	ch := make(chan asyncResult[T], 1)
	go func() {
		value, err := fn(ctx)
		ch <- asyncResult[T]{value: value, err: err}
	}()
	return ch
}

func fetchJSON[T any](ctx context.Context, s *Server, name string, baseURL string) (T, error) {
	var zero T
	cb := s.breakers[name]
	if cb != nil && !cb.Allow() {
		return zero, fmt.Errorf("%s circuit is open", name)
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
		if err != nil {
			return zero, fmt.Errorf("create %s request: %w", name, err)
		}
		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			lastErr = fmt.Errorf("%s returned status %d", name, resp.StatusCode)
			_ = resp.Body.Close()
			continue
		}
		var out T
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			lastErr = err
			_ = resp.Body.Close()
			continue
		}
		_ = resp.Body.Close()
		if cb != nil {
			cb.RecordSuccess()
		}
		return out, nil
	}
	if cb != nil {
		cb.RecordFailure()
	}
	return zero, fmt.Errorf("fetch %s: %w", name, lastErr)
}

func (s *Server) cachedHome() ([]byte, bool) {
	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	if len(s.cache.body) == 0 || time.Now().After(s.cache.expires) {
		return nil, false
	}
	var body mobile.HomeResponse
	if err := json.Unmarshal(s.cache.body, &body); err != nil {
		return nil, false
	}
	body.Meta.Cached = true
	data, err := json.Marshal(body)
	if err != nil {
		return nil, false
	}
	return data, true
}

func (s *Server) storeHome(body []byte) {
	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	s.cache.body = append(s.cache.body[:0], body...)
	s.cache.expires = time.Now().Add(s.cacheTTL)
}

func clientKey(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
