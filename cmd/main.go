package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/scarifel/airbyte-email-notification/config"
	"github.com/scarifel/airbyte-email-notification/internal"
	"github.com/scarifel/airbyte-email-notification/logger"
)

func main() {
	logger.Info("Initalization server...")
	logger.Info(fmt.Sprintf("Server mode: %s", logger.LogLevelString()))

	// загрузка конфигурации
	config, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error loading configuration: %s", err.Error()))
	}

	// инициализация server
	addr := fmt.Sprintf("%s:%s", config.App.Host, config.App.Port)
	server := internal.NewHTTPServer(addr)
	defer server.Close()

	// инициализация smtp
	smtp := internal.NewSMTP(internal.SMTPConfig{
		Host:            config.SMTP.Host,
		Port:            config.SMTP.Port,
		AnonymousAccess: config.SMTP.AnonymousAccess,
		TLS:             config.SMTP.TLS,
		Username:        config.SMTP.Username,
		Password:        config.SMTP.Password,
		MailConfig: internal.MailConfig{
			From:    config.SMTP.From,
			To:      config.SMTP.To,
			Subject: config.SMTP.Subject,
		},
	})

	if err := smtp.Connection(); err != nil {
		logger.Fatal(fmt.Sprintf("Error connection to SMTP server: %s", err.Error()))
	}
	defer smtp.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGTERM)

	// запуск HTTP сервера
	go func() {
		logger.Info(fmt.Sprintf("Starting server on %s", server.HttpServer.Addr))

		if err := server.HttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(fmt.Sprintf("Error starting http server: %s", err.Error()))
		}
	}()

	// отправка сообщений по SMTP
	go func() {
		for message := range server.Messages() {
			if err := smtp.SendMessage(message); err != nil {
				logger.Fatal(fmt.Sprintf("Email sent failed: %s", err.Error()))
			}

			logger.Info(fmt.Sprintf("Email sent success:\r\n"+
				"\tEvent: %s\r\n"+
				"\tStream: %s\r\n"+
				"\tSync start time: %s\r\n"+
				"\tSync end time: %s\r\n"+
				"\tRecord processed: %d\r\n"+
				"\tError message: %s",
				message.Event,
				message.Stream,
				message.SyncStartTime.Format("2006-01-02 15:04:05"),
				message.SyncEndTime.Format("2006-01-02 15:04:05"),
				message.RecordsProcessed,
				message.ErrorMessage),
			)
		}
	}()

	<-interrupt
	logger.Info("Shutting down server...")

	if err := server.HttpServer.Shutdown(context.Background()); err != nil {
		logger.Fatal(fmt.Sprintf("Server shutdown failed: %s", err.Error()))
	}

	logger.Info("Server was stopped")
}
