package kafka

import (
	"context"
	"encoding/json"
	"hookify/internal/models"
	"log/slog"
	"strconv"
	"time"

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
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.Hash{},
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
			WriteTimeout:           10 * time.Second,
			ReadTimeout:            10 * time.Second,
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

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(event.WebhookID, 10)),
		Value: value,
	})
	if err == nil {
		p.log.Info("successfully published event to kafka", "event_id", event.ID, "webhook_id", event.WebhookID)
	}
	return err
}
