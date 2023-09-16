package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Server struct {
	limiter *RateLimiter
}

func NewServer(limiter *RateLimiter) *Server {
	return &Server{
		limiter: limiter,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("XForwarded-For")

	if s.limiter.IsLimited(ip) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	}

	s.limiter.Increment(ip)

	w.Write([]byte("Hello, World!"))
}

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

func extractSubnet(ip string, prefixSizeInBits int) string {
	prefixSizeInBytes := prefixSizeInBits / 8

	parts := strings.Split(ip, ".")

	return strings.Join(parts[:prefixSizeInBytes], ".")
}

func main() {
	prefixSize := 24
	limit := 100
	cooldown := 60 * time.Second
	l := NewRateLimiter(prefixSize, limit, cooldown)
	s := NewServer(l)

	log.Fatal(http.ListenAndServe(":8080", s))
}
