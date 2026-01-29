package models

import (
	"encoding/json"
	"time"
)

type EventStatus string

const (
	EventStatusPending   EventStatus = "pending"
	EventStatusDelivered EventStatus = "delivered"
	EventStatusFailed    EventStatus = "failed"
)

type Webhook struct {
	ID     int64  `json:"id"`
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

type RawEvent struct {
	ID        int64       `json:"id"`
	WebhookID int64       `json:"webhook_id"`
	Payload   string      `json:"payload"`
	Status    EventStatus `json:"status"`
}

type OutboxType string

const (
	OutboxTypePublish  OutboxType = "publish"
	OutboxTypeDelivery OutboxType = "delivery"
)

type OutboxEntry struct {
	ID            int64      `json:"id"`
	Type          OutboxType `json:"type"`
	EventID       int64      `json:"event_id"`
	WebhookID     int64      `json:"webhook_id"`
	Payload       string     `json:"payload"`
	Attempts      int        `json:"attempts"`
	NextAttemptAt time.Time  `json:"next_attempt_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (e *RawEvent) UnmarshalJSON(data []byte) error {
	type rawEventAlias RawEvent

	var v struct {
		rawEventAlias
		HookID int64 `json:"hook_id"`
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*e = RawEvent(v.rawEventAlias)
	if e.WebhookID == 0 && v.HookID != 0 {
		e.WebhookID = v.HookID
	}

	return nil
}
