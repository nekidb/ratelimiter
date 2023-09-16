package server

import (
	"fmt"
	"io"
	"net/http"
)

// Итерфейс лимитера используется для возможности простой замены лимитера, и чтобы можно было мокать
type Limiter interface {
	Increment(ip string)
	IsLimited(ip string) bool
	Reset(subnet string)
}

// Структура сервера содержит роутер для маршрутизации и лимитер для ограчения
type Server struct {
	router  *http.ServeMux
	limiter Limiter
}

func NewServer(limiter Limiter) *Server {
	// Создаем роутер, сервер
	router := http.NewServeMux()
	server := &Server{
		router:  router,
		limiter: limiter,
	}

	// Регистрируем хендлеры
	router.HandleFunc("/", server.defaultHandler)
	router.HandleFunc("/reset", server.resetHandler)

	return server
}

// Хендлер обрабатывает обычные запросы к серверу
func (s Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("XForwarded-For")

	// Если для подсети данного ip превышен лимит запросов, то возвращаем "429 Too Many Requests"
	if s.limiter.IsLimited(ip) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	}

	// Если лимит не превышен, то увеличиваем счетик лимитера и выводим статический контент
	s.limiter.Increment(ip)

	w.Write([]byte("Hello, World!"))

}

// Хендлер обрабатывает запросы на сброс лимита по префиксу
func (s Server) resetHandler(w http.ResponseWriter, r *http.Request) {
	// Вытаскиваем префикс из тела запроса
	subnetBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	defer r.Body.Close()

	subnet := string(subnetBytes)

	// Сбрасываем для префикса его лимит
	s.limiter.Reset(subnet)

	w.Write([]byte(fmt.Sprintf("Limit for %s was reseted", subnet)))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
