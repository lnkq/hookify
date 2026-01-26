package hookify

import "log/slog"

type Service struct {
	log *slog.Logger
}

func New(log *slog.Logger) *Service {
	return &Service{log: log}
}

func (s *Service) CreateWebhook(url string) (hookID int64, secret string, err error) {
	return 1, "secret", nil
}
