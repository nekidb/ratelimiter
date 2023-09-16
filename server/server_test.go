package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nekidb/rate_limiter/ratelimiter"
	"github.com/nekidb/rate_limiter/server"
	"github.com/stretchr/testify/assert"
)

const (
	prefixSize = 24
	limit      = 100
	cooldown   = 1 * time.Second
)

func TestServeHTTP(t *testing.T) {
	limiter := ratelimiter.NewRateLimiter(prefixSize, limit, cooldown)
	server := server.NewServer(limiter)

	t.Run("returns OK when not limited", func(t *testing.T) {
		request := createRequest("123.123.0.1")
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "Hello, World!", recorder.Body.String())
	})

	t.Run("returns error when limited", func(t *testing.T) {
		recorder := httptest.NewRecorder()

		request := createRequest("123.45.67.89")
		for i := 0; i < 32; i++ {
			server.ServeHTTP(recorder, request)
		}

		request = createRequest("123.45.67.1")
		for i := 0; i < 68; i++ {
			server.ServeHTTP(recorder, request)
		}

		recorder = httptest.NewRecorder()
		request = createRequest("123.45.67.111")

		server.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
		assert.Equal(t, "Too Many Requests", recorder.Body.String())
	})
}

func createRequest(ip string) *http.Request {
	request := httptest.NewRequest("GET", "http://localhost:8080/", nil)
	request.Header.Add("XForwarded-For", ip)

	return request
}
