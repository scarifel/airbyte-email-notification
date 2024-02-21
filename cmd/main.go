package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/scarifel/airbyte-email-notification/config"
	"github.com/scarifel/airbyte-email-notification/internal"
)

func main() {
	// load config
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	// инициализация server
	server := internal.NewHTTPServer(":8080")
	defer server.Close()

	// инициализация smtp
	smtp := internal.NewSMTP(internal.SMTPServer{
		Host:            config.SMTP.Host,
		Port:            config.SMTP.Port,
		AnonymousAccess: config.SMTP.AnonymousAccess,
		TLS:             config.SMTP.TLS,
		Username:        config.SMTP.Username,
		Password:        config.SMTP.Password,
		From:            config.SMTP.From,
		To:              config.SMTP.To,
		Subject:         config.SMTP.Subject,
	})

	if err := smtp.Connection(); err != nil {
		log.Fatalf("Error connection to SMTP server: %s", err.Error())
	}
	defer smtp.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGTERM)

	// Отправка сообщений по SMTP
	go func() {
		for message := range server.Messages() {
			if err := smtp.SendMessage(message); err != nil {
				log.Fatalf("Email sent failed: %s", err.Error())
			}

			log.Printf("Email sent successfully. event: %s, stream: %s, start_time: %s, end_time: %s, record_processed: %d \n",
				message.Event,
				message.Stream,
				message.SyncStartTime.Format("2006-01-02 15:04:05"),
				message.SyncEndTime.Format("2006-01-02 15:04:05"),
				message.RecordsProcessed,
			)
		}
	}()

	// Запуск HTTP сервера
	go func() {
		log.Printf("Starting server on %s\n", server.HttpServer.Addr)
		if err := server.HttpServer.ListenAndServe(); err != nil {
			log.Fatalf("Error starting http server: %s", err.Error())
		}
	}()

	<-interrupt
	log.Println("Shutting down server...")

	if err := server.HttpServer.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %s", err.Error())
	}

	log.Println("Server stopped gracefully")
}
