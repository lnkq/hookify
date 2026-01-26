package grpc

import (
	"fmt"
	"log/slog"
	"net"

	hookifygrpc "hookify/internal/grpc"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, webhookAPI hookifygrpc.WebhookAPI, port int) *App {
	gRPCServer := grpc.NewServer()
	hookifygrpc.Register(gRPCServer, webhookAPI)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With(slog.String("op", op))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: failed to listen on port %d: %w", op, a.port, err)
	}

	log.Info("gRPC server is started", slog.Int("port", a.port))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: failed to serve gRPC server: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))
	a.gRPCServer.GracefulStop()
}
