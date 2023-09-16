package main

import (
	"log"
	"net/http"
	"time"

	"github.com/nekidb/rate_limiter/ratelimiter"
	"github.com/nekidb/rate_limiter/server"
)

func main() {
	prefixSize := 24
	limit := 100
	cooldown := 60 * time.Second

	l := ratelimiter.NewRateLimiter(prefixSize, limit, cooldown)
	s := server.NewServer(l)

	log.Fatal(http.ListenAndServe(":8080", s))
}
