package airbytenotificationwebhooksmtp

import "time"

// Message представляет структуру сообщения для отправки
type Message struct {
	Event            string    `json:"event"`
	Stream           string    `json:"stream"`
	SyncStartTime    time.Time `json:"sync_start_time"`
	SyncEndTime      time.Time `json:"sync_end_time"`
	RecordsProcessed uint64    `json:"records_processed"`
	ErrorMessage     string    `json:"error_message"`
}