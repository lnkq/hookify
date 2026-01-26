package grpc

import (
	"context"
	pb "hookify/gen/hookify"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WebhookAPI interface {
	CreateWebhook(ctx context.Context, url string) (hookID int64, secret string, err error)
}

type serverAPI struct {
	pb.UnimplementedHookifyServer
	webhookAPI WebhookAPI
}

func Register(s *grpc.Server, api WebhookAPI) {
	pb.RegisterHookifyServer(s, &serverAPI{webhookAPI: api})
}

func (s *serverAPI) CreateWebhook(ctx context.Context, req *pb.CreateWebhookRequest) (*pb.CreateWebhookResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	_, err := url.ParseRequestURI(req.Url)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	hookID, secret, err := s.webhookAPI.CreateWebhook(ctx, req.Url)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create webhook")
	}

	resp := &pb.CreateWebhookResponse{
		Id:     hookID,
		Secret: secret,
	}

	return resp, nil
}
