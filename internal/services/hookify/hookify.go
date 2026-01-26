package hookify

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hookify/internal/models"
	"log/slog"
)

type Service struct {
	log            *slog.Logger
	webhookSaver   WebhookSaver
	eventSaver     EventSaver
	eventPublisher EventPublisher
}

type WebhookSaver interface {
	SaveWebhook(ctx context.Context, url string, secret string) (int64, error)
}

type EventSaver interface {
	SaveEvent(ctx context.Context, webhookID int64, payload string) (int64, error)
}

type EventPublisher interface {
	PublishEvent(ctx context.Context, event models.RawEvent) error
}

func New(log *slog.Logger, webhookSaver WebhookSaver, eventSaver EventSaver, eventPublisher EventPublisher) *Service {
	return &Service{log: log, webhookSaver: webhookSaver, eventSaver: eventSaver, eventPublisher: eventPublisher}
}

func (s *Service) CreateWebhook(ctx context.Context, url string) (webhookID int64, secret string, err error) {
	secretBytes := make([]byte, 32)
	_, err = rand.Read(secretBytes)
	if err != nil {
		s.log.Error("failed to generate secret", "error", err)
		return 0, "", err
	}
	secret = hex.EncodeToString(secretBytes)

	webhookID, err = s.webhookSaver.SaveWebhook(ctx, url, secret)
	if err != nil {
		s.log.Error("failed to save webhook", "error", err)
		return 0, "", err
	}

	return webhookID, secret, nil
}

func (s *Service) SubmitEvent(ctx context.Context, webhookID int64, payload string, secret string) (eventID int64, err error) {
	eventID, err = s.eventSaver.SaveEvent(ctx, webhookID, payload)
	if err != nil {
		return 0, fmt.Errorf("failed to save event: %w", err)
	}

	// TODO: This design is naive and should be improved in the future.
	rawEvent := models.RawEvent{
		ID:        eventID,
		WebhookID: webhookID,
		Payload:   payload,
	}

	err = s.eventPublisher.PublishEvent(ctx, rawEvent)
	if err != nil {
		s.log.Error("failed to publish event", "error", err)
	}

	return eventID, nil
}
