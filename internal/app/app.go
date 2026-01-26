package app

import (
	"log/slog"

	"hookify/internal/services/hookify"

	grpcapp "hookify/internal/app/grpc"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int) *App {
	hookifyService := hookify.New(log)

	gRPCServer := grpcapp.New(log, hookifyService, grpcPort)

	return &App{
		log:        log,
		GRPCServer: gRPCServer,
	}
}
