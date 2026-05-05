package breaker

import (
	"testing"
	"time"
)

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(2, time.Minute)

	if !cb.Allow() {
		t.Fatal("初期状態でリクエストが許可されませんでした")
	}
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.Allow() {
		t.Fatal("失敗しきい値到達後にリクエストが許可されました")
	}
}

func TestCircuitBreakerHalfOpenAfterCooldown(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	cb.RecordFailure()
	if cb.Allow() {
		t.Fatal("クールダウン前にリクエストが許可されました")
	}
	time.Sleep(15 * time.Millisecond)
	if !cb.Allow() {
		t.Fatal("クールダウン後にリクエストが許可されませんでした")
	}
	cb.RecordSuccess()
	if !cb.Allow() {
		t.Fatal("成功記録後にリクエストが許可されませんでした")
	}
}
