package breaker

import (
	"sync"
	"time"
)

type state int

const (
	closed state = iota
	open
	halfOpen
)

// CircuitBreaker は連続失敗時に下流呼び出しを一時停止します。
type CircuitBreaker struct {
	mu           sync.Mutex
	failures     int
	threshold    int
	cooldown     time.Duration
	openedAt     time.Time
	currentState state
}

// NewCircuitBreaker は失敗しきい値とクールダウンを指定してCircuitBreakerを作成します。
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	if threshold < 1 {
		threshold = 1
	}
	if cooldown <= 0 {
		cooldown = time.Second
	}
	return &CircuitBreaker{threshold: threshold, cooldown: cooldown}
}

// Allow は下流呼び出しを実行できるかを返します。
func (c *CircuitBreaker) Allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentState != open {
		return true
	}
	if time.Since(c.openedAt) < c.cooldown {
		return false
	}
	c.currentState = halfOpen
	return true
}

// RecordSuccess は下流呼び出しの成功を記録します。
func (c *CircuitBreaker) RecordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures = 0
	c.currentState = closed
	c.openedAt = time.Time{}
}

// RecordFailure は下流呼び出しの失敗を記録します。
func (c *CircuitBreaker) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures++
	if c.failures >= c.threshold {
		c.currentState = open
		c.openedAt = time.Now()
	}
}
