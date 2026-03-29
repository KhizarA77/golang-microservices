package ratelimiter

import (
	"sync"
	"time"
)

type IPLimiter struct {
	tokens     int
	maxTokens  int
	lastRefill time.Time
	refillRate time.Duration
}

type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*IPLimiter
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*IPLimiter),
	}
}

func (l *RateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	val, ok := l.limiters[ip]
	if !ok {
		val = &IPLimiter{
			tokens:     5,
			maxTokens:  5,
			lastRefill: time.Now(),
			refillRate: 200 * time.Millisecond,
		}
		l.limiters[ip] = val
	}
	elapsed := time.Since(val.lastRefill)
	newTokens := int(elapsed / val.refillRate)
	if newTokens > 0 {
		val.tokens = min(val.maxTokens, val.tokens+newTokens)
		val.lastRefill = time.Now()
	}
	if val.tokens <= 0 {
		return false
	}
	val.tokens--
	return true
}
