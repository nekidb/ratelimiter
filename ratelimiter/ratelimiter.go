package ratelimiter

import (
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	counter    map[string]int
	lastHits   map[string]time.Time
	prefixSize int
	limit      int
	cooldown   time.Duration
	mu         sync.RWMutex
}

func NewRateLimiter(prefixSize int, limit int, cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		counter:    make(map[string]int),
		lastHits:   make(map[string]time.Time),
		prefixSize: prefixSize,
		limit:      limit,
		cooldown:   cooldown,
	}
}

func (l *RateLimiter) Increment(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	subnet := extractSubnet(ip, l.prefixSize)

	if lastHit, ok := l.lastHits[subnet]; ok && time.Since(lastHit) > l.cooldown {
		l.counter[subnet] = 0
	}

	l.counter[subnet]++
	l.lastHits[subnet] = time.Now()
}

func (l *RateLimiter) IsLimited(ip string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	subnet := extractSubnet(ip, l.prefixSize)

	if lastHit, ok := l.lastHits[subnet]; ok && time.Since(lastHit) > l.cooldown {
		return false
	}

	if l.counter[subnet] >= l.limit {
		return true
	}

	return false
}

func (l *RateLimiter) Reset(subnet string) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	delete(l.counter, subnet)
	delete(l.lastHits, subnet)
}

func extractSubnet(ip string, prefixSizeInBits int) string {
	prefixSizeInBytes := prefixSizeInBits / 8

	parts := strings.Split(ip, ".")

	return strings.Join(parts[:prefixSizeInBytes], ".")
}
