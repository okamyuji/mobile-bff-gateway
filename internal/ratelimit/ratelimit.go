package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	tokens     int
	lastRefill time.Time
}

// Limiter はキーごとのトークンバケット型レート制限です。
type Limiter struct {
	mu       sync.Mutex
	capacity int
	refill   int
	every    time.Duration
	buckets  map[string]bucket
	now      func() time.Time
}

// NewLimiter は指定した容量と補充間隔でLimiterを作成します。
func NewLimiter(capacity int, refill int, every time.Duration) *Limiter {
	if capacity < 1 {
		capacity = 1
	}
	if refill < 0 {
		refill = 0
	}
	if every <= 0 {
		every = time.Second
	}
	return &Limiter{
		capacity: capacity,
		refill:   refill,
		every:    every,
		buckets:  make(map[string]bucket),
		now:      time.Now,
	}
}

// Allow は指定キーのリクエストを許可できるかを返します。
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b := l.buckets[key]
	if b.lastRefill.IsZero() {
		b = bucket{tokens: l.capacity, lastRefill: now}
	}
	if l.refill > 0 {
		elapsed := now.Sub(b.lastRefill)
		steps := int(elapsed / l.every)
		if steps > 0 {
			b.tokens += steps * l.refill
			if b.tokens > l.capacity {
				b.tokens = l.capacity
			}
			b.lastRefill = b.lastRefill.Add(time.Duration(steps) * l.every)
		}
	}
	if b.tokens <= 0 {
		l.buckets[key] = b
		return false
	}
	b.tokens--
	l.buckets[key] = b
	return true
}
