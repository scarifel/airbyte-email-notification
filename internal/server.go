package internal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/scarifel/airbyte-email-notification/logger"
	"github.com/scarifel/airbyte-email-notification/model"
)

type Server struct {
	messages   chan model.Message
	mux        *http.ServeMux
	HttpServer *http.Server
}

func NewHTTPServer(addr string) *Server {
	mux := http.NewServeMux()

	s :=  &Server{
		messages: make(chan model.Message),
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

	var payload model.Message
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		logger.Error(fmt.Sprintf("Failed to decode request body\r\n" + "Body: %s", r.Body))
		return
	}

	s.messages <- payload
	w.WriteHeader(http.StatusOK)
}

// Messages возвращает канал для чтения сообщений
func (s *Server) Messages() <-chan model.Message {
	return s.messages
}

// Close выполняет закрытие канала
func (s *Server) Close() {
	close(s.messages)
}
