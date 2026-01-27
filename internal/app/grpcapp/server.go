package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	"hookify/internal/transport/grpcapi"

	"google.golang.org/grpc"
)

type Server struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, webhookAPI grpcapi.WebhookAPI, port int) *Server {
	gRPCServer := grpc.NewServer()
	grpcapi.Register(gRPCServer, webhookAPI, log)
	return &Server{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (s *Server) Run() error {
	const op = "grpcapp.Run"
	log := s.log.With(slog.String("op", op))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("%s: failed to listen on port %d: %w", op, s.port, err)
	}

	log.Info("gRPC server is started", slog.Int("port", s.port))

	if err := s.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: failed to serve gRPC server: %w", op, err)
	}

	return nil
}

func (s *Server) Stop() {
	const op = "grpcapp.Stop"
	s.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", s.port))
	s.gRPCServer.GracefulStop()
}
