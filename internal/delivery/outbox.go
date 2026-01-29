package delivery

import (
	"context"
	"hookify/internal/models"
	"time"
)

func (s *Service) RunOutboxWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("outbox worker stopped")
			return
		case <-ticker.C:
			if err := s.processOutbox(ctx); err != nil {
				s.log.Error("failed to process outbox", "error", err)
			}
		}
	}
}

func (s *Service) processOutbox(ctx context.Context) error {
	entries, err := s.outboxRepo.GetDueOutboxEntries(ctx, 10)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		rawEvent := models.RawEvent{
			ID:        entry.EventID,
			WebhookID: entry.WebhookID,
			Payload:   entry.Payload,
		}
		err = s.eventPublisher.PublishEvent(ctx, rawEvent)
		if err != nil {
			s.log.Error("failed to publish outbox entry", "error", err)

			nextAttempt := time.Now().Add(time.Duration(entry.Attempts+1) * 5 * time.Second)
			if updateErr := s.outboxRepo.UpdateOutboxEntry(ctx, entry.ID, entry.Attempts+1, nextAttempt); updateErr != nil {
				s.log.Error("failed to update outbox entry", "id", entry.ID, "error", updateErr)
			}
			continue
		}

		if err := s.outboxRepo.DeleteOutboxEntry(ctx, entry.ID); err != nil {
			s.log.Error("failed to delete outbox entry", "id", entry.ID, "error", err)
		}
	}

	return nil
}
