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
	readers []*kafka.Reader
	handler Handler
	workers int
}

func NewConsumer(log *slog.Logger, brokers []string, topic, groupID string, handler Handler, workers int) *Consumer {
	if workers <= 0 {
		workers = 1
	}

	readers := make([]*kafka.Reader, 0, workers)
	for i := 0; i < workers; i++ {
		readers = append(readers, kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MaxBytes: 10e6,
		}))
	}

	return &Consumer{
		log:     log,
		readers: readers,
		handler: handler,
		workers: workers,
	}
}

func (c *Consumer) Close() error {
	var firstErr error
	for _, r := range c.readers {
		if err := r.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (c *Consumer) Run(ctx context.Context) error {
	errCh := make(chan error, len(c.readers))

	for _, r := range c.readers {
		reader := r
		go func() {
			errCh <- c.runReader(ctx, reader)
		}()
	}

	var firstErr error
	for range c.readers {
		if err := <-errCh; err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (c *Consumer) runReader(ctx context.Context, reader *kafka.Reader) error {
	for {
		m, err := reader.FetchMessage(ctx)
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
			if err := reader.CommitMessages(ctx, m); err != nil {
				c.log.Error("failed to commit poison message", "error", err)
			}
			continue
		}

		if err := c.handler.HandleEvent(ctx, event); err != nil {
			c.log.Error("failed to handle event", "error", err)
			continue
		}

		if err := reader.CommitMessages(ctx, m); err != nil {
			c.log.Error("failed to commit message", "error", err)
		}
	}
}
