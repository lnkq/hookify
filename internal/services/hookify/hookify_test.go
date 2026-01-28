package hookify

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"hookify/internal/models"
)

type webhookRepoMock struct {
	saveURL    string
	saveSecret string
	saveID     int64
	saveErr    error

	getWebhook models.Webhook
	getErr     error
}

func (m *webhookRepoMock) SaveWebhook(ctx context.Context, url string, secret string) (int64, error) {
	m.saveURL = url
	m.saveSecret = secret
	return m.saveID, m.saveErr
}

func (m *webhookRepoMock) GetWebhook(ctx context.Context, webhookID int64) (models.Webhook, error) {
	return m.getWebhook, m.getErr
}

type eventSaverMock struct {
	savedWebhookID int64
	savedPayload   string
	id             int64
	err            error
}

func (m *eventSaverMock) SaveEvent(ctx context.Context, webhookID int64, payload string) (int64, error) {
	m.savedWebhookID = webhookID
	m.savedPayload = payload
	return m.id, m.err
}

type publisherMock struct {
	published models.RawEvent
	err       error
	called    bool
}

func (m *publisherMock) PublishEvent(ctx context.Context, event models.RawEvent) error {
	m.published = event
	m.called = true
	return m.err
}

func TestCreateWebhook_GeneratesSecretAndSaves(t *testing.T) {
	repo := &webhookRepoMock{saveID: 123}
	svc := New(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})), repo, &eventSaverMock{}, &publisherMock{})

	id, secret, err := svc.CreateWebhook(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != 123 {
		t.Fatalf("expected id=123, got %d", id)
	}
	if repo.saveURL != "https://example.com" {
		t.Fatalf("expected SaveWebhook url to match, got %q", repo.saveURL)
	}
	if secret == "" {
		t.Fatalf("expected secret to be non-empty")
	}
	if len(secret) != 64 {
		t.Fatalf("expected hex secret length=64, got %d", len(secret))
	}
	if repo.saveSecret != secret {
		t.Fatalf("expected saved secret to match returned secret")
	}
}

func TestSubmitEvent_WebhookNotFound(t *testing.T) {
	repo := &webhookRepoMock{getErr: models.ErrWebhookNotFound}
	svc := New(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})), repo, &eventSaverMock{}, &publisherMock{})

	_, err := svc.SubmitEvent(context.Background(), 1, `{}`, "x")
	if !errors.Is(err, models.ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound, got %v", err)
	}
}

func TestSubmitEvent_InvalidSecret(t *testing.T) {
	repo := &webhookRepoMock{getWebhook: models.Webhook{ID: 1, URL: "https://example.com", Secret: "expected"}}
	svc := New(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})), repo, &eventSaverMock{}, &publisherMock{})

	_, err := svc.SubmitEvent(context.Background(), 1, `{}`, "wrong")
	if !errors.Is(err, ErrInvalidWebhookSecret) {
		t.Fatalf("expected ErrInvalidWebhookSecret, got %v", err)
	}
}

func TestSubmitEvent_PublishesPendingEvent(t *testing.T) {
	repo := &webhookRepoMock{getWebhook: models.Webhook{ID: 7, URL: "https://example.com", Secret: "s"}}
	saver := &eventSaverMock{id: 99}
	pub := &publisherMock{}
	svc := New(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})), repo, saver, pub)

	id, err := svc.SubmitEvent(context.Background(), 7, `{"a":1}`, "s")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != 99 {
		t.Fatalf("expected event id=99, got %d", id)
	}
	if !pub.called {
		t.Fatalf("expected publisher to be called")
	}
	if pub.published.ID != 99 || pub.published.WebhookID != 7 || pub.published.Payload != `{"a":1}` || pub.published.Status != models.EventStatusPending {
		t.Fatalf("unexpected published event: %#v", pub.published)
	}
}

func TestSubmitEvent_PublishErrorDoesNotFailRequest(t *testing.T) {
	repo := &webhookRepoMock{getWebhook: models.Webhook{ID: 7, URL: "https://example.com", Secret: "s"}}
	saver := &eventSaverMock{id: 99}
	pub := &publisherMock{err: errors.New("kafka down")}
	svc := New(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})), repo, saver, pub)

	id, err := svc.SubmitEvent(context.Background(), 7, `{}`, "s")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != 99 {
		t.Fatalf("expected event id=99, got %d", id)
	}
}
