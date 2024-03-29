package internal

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/scarifel/airbyte-email-notification/model"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type MailConfig struct {
	From    string
	To      []string
	Subject string
}

type SMTPConfig struct {
	Host            string
	Port            int
	AnonymousAccess bool
	TLS             bool
	Username        string
	Password        string
	MailConfig      MailConfig
}

type SMTP struct {
	config SMTPConfig
	conn   *smtp.Client
}

func NewSMTP(config SMTPConfig) *SMTP {
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

// connect makes a connection to the SMTP server
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

// checkEmailAddresses checks the contact details of the sender and the recipient(s) of the message
func (s *SMTP) checkEmailAddresses() error {
	if s.config.MailConfig.From == "" {
		return errors.New("sender email cannot be empty")
	}

	if len(s.config.MailConfig.To) == 0 {
		return errors.New("recipient email address cannot be empty")
	}

	return nil
}

// Close closes the current connection
func (s *SMTP) Close() error {
	return s.conn.Close()
}

// SendMessage performs the sending of a message
func (s *SMTP) SendMessage(message model.Message) error {
	// sender
	if err := s.conn.Mail(s.config.MailConfig.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// receiver
	if err := s.conn.Rcpt(strings.Join(s.config.MailConfig.To, ",")); err != nil {
		return fmt.Errorf("failed to add recipients: %w", err)
	}
	w, err := s.conn.Data()
	if err != nil {
		return fmt.Errorf("failed to initiate message data: %w", err)
	}
	defer w.Close()

	buildedMessage := s.buildMail(message)

	if _, err = w.Write(buildedMessage); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}
	return nil
}

func (s *SMTP) buildMail(m model.Message) []byte {
	caserTitle := cases.Title(language.English)
	caserUpper := cases.Upper(language.English)

	subject := fmt.Sprintf("Subject: [%s] %s %s syncronization\n",
		s.config.MailConfig.Subject, caserTitle.String(m.Stream), caserUpper.String(m.Event))

	body := fmt.Sprintf("Event: %s\n"+
		"Stream: %s\n"+
		"Sync Start: %s\n"+
		"Sync End: %s\n"+
		"Records Processed: %d\n",
		m.Event, m.Stream, m.SyncStartTime, m.SyncEndTime, m.RecordsProcessed,
	)

	if m.ErrorMessage != "" {
		body += fmt.Sprintf("Error Message: %s\n", m.ErrorMessage)
	}

	return []byte(subject + "\n" + body)
}
