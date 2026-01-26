package delivery

import (
	"bytes"
	"context"
	"fmt"
	"hookify/internal/models"
	"log/slog"
	"net/http"
)

type Service struct {
	log             *slog.Logger
	webhookProvider WebhookProvider
}

type WebhookProvider interface {
	GetWebhook(ctx context.Context, hookID int64) (models.Webhook, error)
}

func New(log *slog.Logger, webhookProvider WebhookProvider) *Service {
	return &Service{
		log:             log,
		webhookProvider: webhookProvider,
	}
}

func (s *Service) HandleEvent(ctx context.Context, event models.RawEvent) error {
	s.log.Info("handling event", "event_id", event.ID, "hook_id", event.HookID)

	// TODO: Use worker pool
	webhook, err := s.webhookProvider.GetWebhook(ctx, event.HookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	err = s.sendRequest(ctx, webhook.URL, webhook.Secret, event.Payload)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	s.log.Info("event handled successfully", "event_id", event.ID, "hook_id", event.HookID)

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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response: %d", resp.StatusCode)
	}

	return nil
}
