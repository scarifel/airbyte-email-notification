package model

import "time"

// Message представляет структуру сообщения для отправки
type Message struct {
	Event            string    `json:"event" validate:"required"`
	Stream           string    `json:"stream" validate:"required"` 
	SyncStartTime    time.Time `json:"sync_start_time" validate:"required"`
	SyncEndTime      time.Time `json:"sync_end_time" validate:"required"`
	RecordsProcessed uint64    `json:"records_processed"`
	ErrorMessage     string    `json:"error_message"`
}