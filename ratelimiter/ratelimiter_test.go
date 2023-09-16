package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/nekidb/rate_limiter/ratelimiter"
)

const (
	prefixSize = 24
	limit      = 100
	cooldown   = 1 * time.Second
)

func TestRateLimiter(t *testing.T) {
	limiter := ratelimiter.NewRateLimiter(prefixSize, limit, cooldown)

	ip := "123.123.0.1"

	if limiter.IsLimited(ip) {
		t.Fatal("Has not to be limited")
	}

	for i := 0; i < 100; i++ {
		limiter.Increment(ip)
	}

	if !limiter.IsLimited(ip) {
		t.Fatal("Has to be limited")
	}

	time.Sleep(cooldown + 1*time.Second)

	if limiter.IsLimited(ip) {
		t.Fatal("Has not to be limited")
	}
}
