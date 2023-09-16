package server

import (
	"log"
	"net/http"
)

type Limiter interface {
	Increment(ip string)
	IsLimited(ip string) bool
	Reset(subnet string)
}

type Server struct {
	router  *http.ServeMux
	limiter Limiter
}

func NewServer(limiter Limiter) *Server {
	router := http.NewServeMux()

	server := &Server{
		router:  router,
		limiter: limiter,
	}

	router.HandleFunc("/", server.defaultHandler)
	router.HandleFunc("/reset", server.resetHandler)

	return server
}

func (s Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving HTTP")
	ip := r.Header.Get("XForwarded-For")

	if s.limiter.IsLimited(ip) {
		log.Println("limited hit")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	}

	s.limiter.Increment(ip)

	w.Write([]byte("Hello, World!"))

}

func (s Server) resetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving HTTP")
	ip := r.Header.Get("XForwarded-For")

	if s.limiter.IsLimited(ip) {
		log.Println("limited hit")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	}

	s.limiter.Increment(ip)

	w.Write([]byte("Hello, World!"))

}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
