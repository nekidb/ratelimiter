package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
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
	subnet := extractSubnet(ip, 24)

	if s.limiter.IsLimited(subnet) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too many requests"))
		return
	}

	s.limiter.Increment(subnet)

	w.Write([]byte("Hello, World!"))
}

type RateLimiter struct {
	counter map[string]int
	mu      sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		counter: make(map[string]int),
	}
}

func (l *RateLimiter) Increment(subnet string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counter[subnet]++
}

func (l *RateLimiter) IsLimited(subnet string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.counter[subnet] >= 100 {
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
	l := NewRateLimiter()
	s := NewServer(l)

	log.Fatal(http.ListenAndServe(":8080", s))
}
