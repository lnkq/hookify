package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hookify/internal/models"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveEventWithOutbox(ctx context.Context, webhookID int64, payload string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var eventID int64
	err = tx.QueryRowContext(ctx, "INSERT INTO events(webhook_id, payload, status) VALUES($1, $2, $3) RETURNING id", webhookID, payload, models.EventStatusPending).Scan(&eventID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert event: %w", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO outbox(event_id, webhook_id, payload, attempts, next_attempt_at) VALUES($1, $2, $3, 0, NOW())", eventID, webhookID, payload)
	if err != nil {
		return 0, fmt.Errorf("failed to insert outbox entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return eventID, nil
}

func (s *Storage) GetDueOutboxEntries(ctx context.Context, limit int) ([]models.OutboxEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, event_id, webhook_id, payload, attempts, next_attempt_at, created_at 
		FROM outbox 
		WHERE next_attempt_at <= NOW() 
		ORDER BY next_attempt_at ASC, id ASC 
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox: %w", err)
	}
	defer rows.Close()

	var entries []models.OutboxEntry
	for rows.Next() {
		var e models.OutboxEntry
		if err := rows.Scan(&e.ID, &e.EventID, &e.WebhookID, &e.Payload, &e.Attempts, &e.NextAttemptAt, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan outbox entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *Storage) UpdateOutboxEntry(ctx context.Context, id int64, attempts int, nextAttemptAt time.Time) error {
	_, err := s.db.ExecContext(ctx, "UPDATE outbox SET attempts=$1, next_attempt_at=$2 WHERE id=$3", attempts, nextAttemptAt, id)
	if err != nil {
		return fmt.Errorf("failed to update outbox entry: %w", err)
	}
	return nil
}

func (s *Storage) DeleteOutboxEntry(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM outbox WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("failed to delete outbox entry: %w", err)
	}
	return nil
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

func (s *Storage) UpdateEventStatus(ctx context.Context, eventID int64, status models.EventStatus) error {
	_, err := s.db.ExecContext(ctx, "UPDATE events SET status=$1 WHERE id=$2", status, eventID)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	return nil
}
