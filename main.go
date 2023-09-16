package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nekidb/rate_limiter/ratelimiter"
	"github.com/nekidb/rate_limiter/server"
)

func main() {
	prefixSize, err := strconv.Atoi(os.Getenv("PREFIX_SIZE"))
	if err != nil {
		log.Fatal("can not convert: ", err)
	}

	limit, err := strconv.Atoi(os.Getenv("RATE_LIMIT"))
	if err != nil {
		log.Fatal("can not convert: ", err)
	}

	cooldown, err := time.ParseDuration(os.Getenv("COOLDOWN"))
	if err != nil {
		log.Fatal("can not convert: ", err)
	}

	l := ratelimiter.NewRateLimiter(prefixSize, limit, cooldown)
	s := server.NewServer(l)

	log.Fatal(http.ListenAndServe(":8080", s))
}
