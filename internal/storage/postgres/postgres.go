package postgres

import (
	"context"
	"database/sql"
	"fmt"

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
	err := s.db.QueryRowContext(ctx, "INSERT INTO webhooks(url, secret) VALUES($1, $2) RETURNING hook_id", url, secret).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert webhook: %w", err)
	}

	return id, nil
}
