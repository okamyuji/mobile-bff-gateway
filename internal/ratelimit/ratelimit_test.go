package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterRejectsWhenBucketIsEmpty(t *testing.T) {
	t.Parallel()

	limiter := NewLimiter(2, 0, time.Second)

	if !limiter.Allow("client-a") {
		t.Fatal("1回目のリクエストが許可されませんでした")
	}
	if !limiter.Allow("client-a") {
		t.Fatal("2回目のリクエストが許可されませんでした")
	}
	if limiter.Allow("client-a") {
		t.Fatal("空のバケットでリクエストが許可されました")
	}
}

func TestLimiterRefillsTokens(t *testing.T) {
	t.Parallel()

	limiter := NewLimiter(1, 1, 10*time.Millisecond)

	if !limiter.Allow("client-a") {
		t.Fatal("初回リクエストが許可されませんでした")
	}
	if limiter.Allow("client-a") {
		t.Fatal("補充前にリクエストが許可されました")
	}
	time.Sleep(15 * time.Millisecond)
	if !limiter.Allow("client-a") {
		t.Fatal("補充後にリクエストが許可されませんでした")
	}
}
