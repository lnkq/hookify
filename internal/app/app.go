package app

import (
	"log/slog"

	"hookify/internal/services/hookify"
	"hookify/internal/storage/postgres"

	grpcapp "hookify/internal/app/grpc"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int) *App {
	storage, err := postgres.New()
	if err != nil {
		log.Error("failed to create postgres storage", "error", err)
		return nil
	}

	hookifyService := hookify.New(log, storage)

	gRPCServer := grpcapp.New(log, hookifyService, grpcPort)

	return &App{
		log:        log,
		GRPCServer: gRPCServer,
	}
}
