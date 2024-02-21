package airbyteemailnotification

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	messages   chan Message
	mux        *http.ServeMux
	HttpServer *http.Server
}

func NewHTTPServer(addr string) *Server {
	mux := http.NewServeMux()

	s :=  &Server{
		messages: make(chan Message),
		mux:      mux,
		HttpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	s.mux.HandleFunc("/send_report", s.handlerMessages)

	return s
}

func (s *Server) handlerMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload Message
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	s.messages <- payload
	w.WriteHeader(http.StatusOK)
}

// Messages возвращает канал для чтения сообщений
func (s *Server) Messages() <-chan Message {
	return s.messages
}

// Close выполняет закрытие канала
func (s *Server) Close() {
	close(s.messages)
}
