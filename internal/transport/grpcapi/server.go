package grpcapi

import (
	"context"
	"log/slog"
	"net/url"

	pb "hookify/gen/hookify"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WebhookAPI interface {
	CreateWebhook(ctx context.Context, url string) (webhookID int64, secret string, err error)
	SubmitEvent(ctx context.Context, webhookID int64, payload string, secret string) (eventID int64, err error)
}

type serverAPI struct {
	pb.UnimplementedHookifyServer
	webhookAPI WebhookAPI
	log        *slog.Logger
}

func Register(s *grpc.Server, api WebhookAPI, log *slog.Logger) {
	pb.RegisterHookifyServer(s, &serverAPI{webhookAPI: api, log: log})
}

func (s *serverAPI) CreateWebhook(ctx context.Context, req *pb.CreateWebhookRequest) (*pb.CreateWebhookResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	_, err := url.ParseRequestURI(req.Url)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	webhookID, secret, err := s.webhookAPI.CreateWebhook(ctx, req.Url)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create webhook")
	}

	resp := &pb.CreateWebhookResponse{
		WebhookId: webhookID,
		Secret:    secret,
	}

	return resp, nil
}

func (s *serverAPI) SubmitEvent(ctx context.Context, req *pb.SubmitEventRequest) (*pb.SubmitEventResponse, error) {
	if req.Secret == "" {
		return nil, status.Error(codes.InvalidArgument, "secret is required")
	}

	eventID, err := s.webhookAPI.SubmitEvent(ctx, req.WebhookId, req.Payload, req.Secret)
	if err != nil {
		s.log.Error("failed to submit event", "error", err)
		return nil, status.Error(codes.Internal, "failed to submit event")
	}

	resp := &pb.SubmitEventResponse{
		EventId: eventID,
		Created: true,
	}

	return resp, nil
}
