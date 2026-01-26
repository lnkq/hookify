package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"hookify/internal/models"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	log *slog.Logger
}

func New(log *slog.Logger) *Publisher {
	return &Publisher{log: log}
}

func (p *Publisher) PublishEvent(context context.Context, event models.RawEvent) error {
	topic := "webhook.events.raw"
	partition := 0

	conn, err := kafka.DialLeader(context, "tcp", "localhost:9092", topic, partition)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}

	j, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", event.ID)),
		Value: j,
	})

	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	if err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}
