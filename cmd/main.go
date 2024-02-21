package main

import (
	"context"
	"fmt"
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
	server := internal.NewHTTPServer(":8080")
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

	// Отправка сообщений по SMTP
	go func() {
		for message := range server.Messages() {
			if err := smtp.SendMessage(message); err != nil {
				logger.Fatal(fmt.Sprintf("Email sent failed: %s", err.Error()))
			}

			logger.Info(fmt.Sprintf("Email sent success." +
			    "Event: %s, Stream: %s," +
				"Sync start time: %s, Sync end time: %s, record_processed: %d",
				message.Event,
				message.Stream,
				message.SyncStartTime.Format("2006-01-02 15:04:05"),
				message.SyncEndTime.Format("2006-01-02 15:04:05"),
				message.RecordsProcessed))
		}
	}()

	// Запуск HTTP сервера
	go func() {
		logger.Info(fmt.Sprintf("Starting server on %s", server.HttpServer.Addr))
		if err := server.HttpServer.ListenAndServe(); err != nil {
			logger.Fatal(fmt.Sprintf("Error starting http server: %s", err.Error()))
		}
	}()

	<-interrupt
	logger.Info("Shutting down server...")

	if err := server.HttpServer.Shutdown(context.Background()); err != nil {
		logger.Fatal(fmt.Sprintf("Server shutdown failed: %s", err.Error()))
	}

	logger.Info("Server was stopped")
}
