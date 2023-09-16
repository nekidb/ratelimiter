package server

import "net/http"

type Limiter interface {
	Increment(ip string)
	IsLimited(ip string) bool
}

type Server struct {
	limiter Limiter
}

func NewServer(limiter Limiter) *Server {
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
