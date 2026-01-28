package grpcapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	pb "hookify/gen/hookify"
	"hookify/internal/models"
	"hookify/internal/services/hookify"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type apiMock struct {
	createID     int64
	createSecret string
	createErr    error
	createURL    string

	submitID      int64
	submitErr     error
	submitHookID  int64
	submitSecret  string
	submitPayload string
}

func (m *apiMock) CreateWebhook(ctx context.Context, url string) (int64, string, error) {
	m.createURL = url
	return m.createID, m.createSecret, m.createErr
}

func (m *apiMock) SubmitEvent(ctx context.Context, webhookID int64, payload string, secret string) (int64, error) {
	m.submitHookID = webhookID
	m.submitPayload = payload
	m.submitSecret = secret
	return m.submitID, m.submitErr
}

func TestCreateWebhook_EmptyURL(t *testing.T) {
	s := &serverAPI{webhookAPI: &apiMock{}, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	_, err := s.CreateWebhook(context.Background(), &pb.CreateWebhookRequest{Url: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestCreateWebhook_InvalidURL(t *testing.T) {
	s := &serverAPI{webhookAPI: &apiMock{}, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	_, err := s.CreateWebhook(context.Background(), &pb.CreateWebhookRequest{Url: "://bad"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestCreateWebhook_OK(t *testing.T) {
	api := &apiMock{createID: 10, createSecret: "sec"}
	s := &serverAPI{webhookAPI: api, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	resp, err := s.CreateWebhook(context.Background(), &pb.CreateWebhookRequest{Url: "https://example.com"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.WebhookId != 10 || resp.Secret != "sec" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if api.createURL != "https://example.com" {
		t.Fatalf("expected createURL to be passed through, got %q", api.createURL)
	}
}

func TestSubmitEvent_EmptySecret(t *testing.T) {
	s := &serverAPI{webhookAPI: &apiMock{}, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	_, err := s.SubmitEvent(context.Background(), &pb.SubmitEventRequest{WebhookId: 1, Payload: `{}`, Secret: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestSubmitEvent_WebhookNotFound(t *testing.T) {
	api := &apiMock{submitErr: models.ErrWebhookNotFound}
	s := &serverAPI{webhookAPI: api, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	_, err := s.SubmitEvent(context.Background(), &pb.SubmitEventRequest{WebhookId: 1, Payload: `{}`, Secret: "s"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}

func TestSubmitEvent_InvalidWebhookSecret(t *testing.T) {
	api := &apiMock{submitErr: hookify.ErrInvalidWebhookSecret}
	s := &serverAPI{webhookAPI: api, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	_, err := s.SubmitEvent(context.Background(), &pb.SubmitEventRequest{WebhookId: 1, Payload: `{}`, Secret: "s"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestSubmitEvent_InternalError(t *testing.T) {
	api := &apiMock{submitErr: errors.New("boom")}
	s := &serverAPI{webhookAPI: api, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	_, err := s.SubmitEvent(context.Background(), &pb.SubmitEventRequest{WebhookId: 1, Payload: `{}`, Secret: "s"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", status.Code(err))
	}
}

func TestSubmitEvent_OK(t *testing.T) {
	api := &apiMock{submitID: 55}
	s := &serverAPI{webhookAPI: api, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	resp, err := s.SubmitEvent(context.Background(), &pb.SubmitEventRequest{WebhookId: 7, Payload: `{"a":1}`, Secret: "sec"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.EventId != 55 || !resp.Created {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if api.submitHookID != 7 || api.submitSecret != "sec" || api.submitPayload != `{"a":1}` {
		t.Fatalf("unexpected args: hookID=%d secret=%q payload=%q", api.submitHookID, api.submitSecret, api.submitPayload)
	}
}
