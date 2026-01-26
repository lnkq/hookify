package hookify

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
)

type Service struct {
	log          *slog.Logger
	webhookSaver WebhookSaver
}

type WebhookSaver interface {
	SaveWebhook(context context.Context, url string, secret string) (int64, error)
}

func New(log *slog.Logger, webhookSaver WebhookSaver) *Service {
	return &Service{log: log, webhookSaver: webhookSaver}
}

func (s *Service) CreateWebhook(context context.Context, url string) (hookID int64, secret string, err error) {
	secretBytes := make([]byte, 32)
	_, err = rand.Read(secretBytes)
	if err != nil {
		s.log.Error("failed to generate secret", "error", err)
		return 0, "", err
	}
	secret = hex.EncodeToString(secretBytes)

	hookID, err = s.webhookSaver.SaveWebhook(context, url, secret)
	if err != nil {
		s.log.Error("failed to save webhook", "error", err)
		return 0, "", err
	}

	return hookID, secret, nil
}
