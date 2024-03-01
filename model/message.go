package model

import (
	"time"
	"github.com/go-playground/validator/v10"
)

// Message represents the structure of the message received from the webhook Airbyte
type Message struct {
	Event            string    `json:"event" validate:"required"`
	Stream           string    `json:"stream" validate:"required"` 
	SyncStartTime    time.Time `json:"sync_start_time" validate:"required"`
	SyncEndTime      time.Time `json:"sync_end_time" validate:"required"`
	RecordsProcessed uint64    `json:"records_processed"`
	ErrorMessage     string    `json:"error_message"`
}

func (m Message) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}