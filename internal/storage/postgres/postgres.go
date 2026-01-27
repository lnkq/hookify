package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hookify/internal/models"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New() (*Storage, error) {
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=hookify sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveWebhook(ctx context.Context, url string, secret string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, "INSERT INTO webhooks(url, secret) VALUES($1, $2) RETURNING id", url, secret).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert webhook: %w", err)
	}

	return id, nil
}

func (s *Storage) GetWebhook(ctx context.Context, webhookID int64) (models.Webhook, error) {
	var webhook models.Webhook
	err := s.db.QueryRowContext(ctx, "SELECT id, url, secret FROM webhooks WHERE id=$1", webhookID).Scan(&webhook.ID, &webhook.URL, &webhook.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Webhook{}, models.ErrWebhookNotFound
		}
		return models.Webhook{}, fmt.Errorf("failed to get webhook: %w", err)
	}

	return webhook, nil
}

func (s *Storage) SaveEvent(ctx context.Context, webhookID int64, payload string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, "INSERT INTO events(webhook_id, payload) VALUES($1, $2) RETURNING id", webhookID, payload).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert event: %w", err)
	}

	return id, nil
}
