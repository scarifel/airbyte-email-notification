package airbytenotificationwebhooksmtp

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPServer struct {
	Host            string
	Port            int
	AnonymousAccess bool
	TLS             bool
	Username        string
	Password        string
	From            string
	To              []string
	Subject         string
}

type SMTP struct {
	config SMTPServer
	conn   *smtp.Client
}

func NewSMTP(config SMTPServer) *SMTP {
	return &SMTP{
		config: config,
	}
}

// Connection реализует подключение к SMTP серверу с проверкой типа подключения
func (s *SMTP) Connection() error {
	if !s.config.AnonymousAccess {
		return s.connectionWithAuth()
	}
	return s.connect()
}

// connectionWithAuth выполняет подключение к SMTP-серверу с авторизацией
func (s *SMTP) connectionWithAuth() error {
	if err := s.connect(); err != nil {
		return err
	}

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	if err := s.conn.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}

// connect выполняет подключение к SMTP-серверу
func (s *SMTP) connect() error {
	if err := s.checkEmailAddresses(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to the SMTP server: %w", err)
	}

	if s.config.TLS {
		if err := conn.StartTLS(&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         addr,
		}); err != nil {
			return fmt.Errorf("failed to start TLS connection: %w", err)
		}
	}

	s.conn = conn
	return nil
}

// checkEmailAddresses выполняет проверку контактных данных отправителя и получателя(ей) сообщения
func (s *SMTP) checkEmailAddresses() error {
	if s.config.From == "" {
		return errors.New("sender email cannot be empty")
	}

	if len(s.config.To) == 0 {
		return errors.New("recipient email address cannot be empty")
	}

	return nil
}

// Close выполняет закрытие текущего подключения
func (s *SMTP) Close() error {
	return s.conn.Close()
}

// SendMessage выполняет отправку сообщения
func (s *SMTP) SendMessage(message Message) error {
	// отправитель
	if err := s.conn.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// получатель
	if err := s.conn.Rcpt(strings.Join(s.config.To, ",")); err != nil {
		return fmt.Errorf("failed to add recipients: %w", err)
	}
	w, err := s.conn.Data()
	if err != nil {
		return fmt.Errorf("failed to initiate message data: %w", err)
	}
	defer w.Close()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encoding message to bytes: %w", err)
	}

	if _, err = w.Write(messageBytes); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}
	return nil
}
