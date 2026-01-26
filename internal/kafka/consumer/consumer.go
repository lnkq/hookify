package consumer

import (
	"context"
	"encoding/json"
	"hookify/internal/models"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type HandleEvent interface {
	HandleEvent(ctx context.Context, event models.RawEvent) error
}

type Consumer struct {
	log         *slog.Logger
	reader      *kafka.Reader
	handleEvent HandleEvent
}

func New(log *slog.Logger, handleEvent HandleEvent) *Consumer {
	return &Consumer{
		log: log,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{"localhost:9092"},
			Topic:     "webhook.events.raw",
			GroupID:   "consumer-group-1",
			Partition: 0,
			MaxBytes:  10e6,
		}),
		handleEvent: handleEvent,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.log.Error("failed to fetch message", "error", err)
			continue
		}

		c.log.Debug("received message", "key", string(m.Key), "value", string(m.Value))

		var event models.RawEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			c.log.Error("failed to unmarshal message", "error", err)
			continue
		}

		if err := c.handleEvent.HandleEvent(ctx, event); err != nil {
			c.log.Error("failed to handle event", "error", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.log.Error("failed to commit message", "error", err)
		}
	}
}
