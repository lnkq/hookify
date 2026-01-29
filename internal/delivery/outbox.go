package delivery

import (
	"context"
	"fmt"
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
		var processErr error

		switch entry.Type {
		case models.OutboxTypePublish:
			rawEvent := models.RawEvent{
				ID:        entry.EventID,
				WebhookID: entry.WebhookID,
				Payload:   entry.Payload,
			}
			processErr = s.eventPublisher.PublishEvent(ctx, rawEvent)
		case models.OutboxTypeDelivery:
			webhook, err := s.webhookProvider.GetWebhook(ctx, entry.WebhookID)
			if err != nil {
				processErr = err
			} else {
				processErr = s.sendRequest(ctx, webhook.URL, webhook.Secret, entry.Payload)
				if processErr == nil {
					if err := s.eventStatusUpdater.UpdateEventStatus(ctx, entry.EventID, models.EventStatusDelivered); err != nil {
						processErr = fmt.Errorf("failed to update event status: %w", err)
					}
				}
			}
		}

		if processErr != nil {
			s.log.Error("failed to process outbox entry", "id", entry.ID, "type", entry.Type, "error", processErr)

			if time.Since(entry.CreatedAt) > 24*time.Hour {
				s.log.Error("outbox entry exceeded 24h deadline, dropping", "id", entry.ID)
				if err := s.eventStatusUpdater.UpdateEventStatus(ctx, entry.EventID, models.EventStatusFailed); err != nil {
					s.log.Error("failed to mark event as failed", "event_id", entry.EventID, "error", err)
				}
				if err := s.outboxRepo.DeleteOutboxEntry(ctx, entry.ID); err != nil {
					s.log.Error("failed to delete expired outbox entry", "id", entry.ID, "error", err)
				}
				continue
			}

			nextAttempts := entry.Attempts + 1
			nextAttemptAt := time.Now().Add(time.Duration(nextAttempts) * 5 * time.Second)

			if updateErr := s.outboxRepo.UpdateOutboxEntry(ctx, entry.ID, nextAttempts, nextAttemptAt); updateErr != nil {
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
