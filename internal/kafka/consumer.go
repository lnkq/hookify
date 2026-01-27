package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"hookify/internal/models"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Handler interface {
	HandleEvent(ctx context.Context, event models.RawEvent) error
}

type Consumer struct {
	log     *slog.Logger
	reader  *kafka.Reader
	handler Handler
}

func NewConsumer(log *slog.Logger, brokers []string, topic, groupID string, handler Handler) *Consumer {
	return &Consumer{
		log: log,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MaxBytes: 10e6,
		}),
		handler: handler,
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			c.log.Error("failed to fetch message", "error", err)
			time.Sleep(time.Second)
			continue
		}

		var event models.RawEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			c.log.Error("failed to unmarshal message", "error", err)
			if err := c.reader.CommitMessages(ctx, m); err != nil {
				c.log.Error("failed to commit poison message", "error", err)
			}
			continue
		}

		if err := c.handler.HandleEvent(ctx, event); err != nil {
			c.log.Error("failed to handle event", "error", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.log.Error("failed to commit message", "error", err)
		}
	}
}
