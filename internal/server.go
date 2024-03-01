package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	s := &Server{
		messages: make(chan model.Message),
		mux:      mux,
		HttpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	s.mux.HandleFunc("/webhook", s.handlerMessages)

	return s
}

func (s *Server) handlerMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusInternalServerError)
		return
	}

	logger.Debug(fmt.Sprintf("Request body:\r\n%s", requestBody))

	body := bytes.NewReader(requestBody)
	var payload model.Message
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		logger.Error(fmt.Sprintf("Failed to decode request body\r\n"+"\tBody: %s", r.Body))
		http.Error(w, "Failed to decode request body", http.StatusOK)
		return 
	}

	if payload.Validate() == nil {
		s.messages <- payload
	}

	w.WriteHeader(http.StatusOK)
}

// Messages returns a channel for reading messages
func (s *Server) Messages() <-chan model.Message {
	return s.messages
}

// Close performs closing of the channel
func (s *Server) Close() {
	close(s.messages)
}
