package delivery

import (
	"bytes"
	"context"
	"fmt"
	"hookify/internal/models"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Service struct {
	log                *slog.Logger
	webhookProvider    WebhookProvider
	eventStatusUpdater EventStatusUpdater
	outboxRepo         OutboxRepository
	eventPublisher     EventPublisher
	httpClient         *http.Client
}

type WebhookProvider interface {
	GetWebhook(ctx context.Context, webhookID int64) (models.Webhook, error)
}

type EventStatusUpdater interface {
	UpdateEventStatus(ctx context.Context, eventID int64, status models.EventStatus) error
}

type OutboxRepository interface {
	GetDueOutboxEntries(ctx context.Context, limit int) ([]models.OutboxEntry, error)
	UpdateOutboxEntry(ctx context.Context, id int64, attempts int, nextAttemptAt time.Time) error
	DeleteOutboxEntry(ctx context.Context, id int64) error
}

type EventPublisher interface {
	PublishEvent(ctx context.Context, event models.RawEvent) error
}

func New(log *slog.Logger, webhookProvider WebhookProvider, eventStatusUpdater EventStatusUpdater, outboxRepo OutboxRepository, eventPublisher EventPublisher) *Service {
	return &Service{
		log:                log,
		webhookProvider:    webhookProvider,
		eventStatusUpdater: eventStatusUpdater,
		outboxRepo:         outboxRepo,
		eventPublisher:     eventPublisher,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *Service) HandleEvent(ctx context.Context, event models.RawEvent) error {
	s.log.Info("handling event", "event_id", event.ID, "webhook_id", event.WebhookID)

	webhook, err := s.webhookProvider.GetWebhook(ctx, event.WebhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	err = s.sendRequest(ctx, webhook.URL, webhook.Secret, event.Payload)
	if err != nil {
		if updateErr := s.eventStatusUpdater.UpdateEventStatus(ctx, event.ID, models.EventStatusFailed); updateErr != nil {
			s.log.Error("failed to update event status to failed", "error", updateErr)
		}
		return fmt.Errorf("failed to send request: %w", err)
	}

	if err := s.eventStatusUpdater.UpdateEventStatus(ctx, event.ID, models.EventStatusDelivered); err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	s.log.Info("event handled successfully", "event_id", event.ID, "webhook_id", event.WebhookID)

	return nil
}

func (s *Service) sendRequest(ctx context.Context, url, secret, payload string) error {
	r := bytes.NewReader([]byte(payload))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("X-Secret", secret)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		if err := resp.Body.Close(); err != nil {
			s.log.Error("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response: %d", resp.StatusCode)
	}

	return nil
}
