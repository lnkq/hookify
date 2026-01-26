package kafka

import (
	"context"
	"encoding/json"
	"hookify/internal/models"
	"log/slog"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	log    *slog.Logger
	writer *kafka.Writer
}

func NewProducer(log *slog.Logger, brokers []string, topic string) *Producer {
	return &Producer{
		log: log,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.Hash{},
		},
	}
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func (p *Producer) PublishEvent(ctx context.Context, event models.RawEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(event.WebhookID, 10)),
		Value: value,
	})
}
