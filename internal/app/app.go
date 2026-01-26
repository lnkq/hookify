package app

import (
	"log/slog"

	"hookify/internal/delivery"
	"hookify/internal/kafka/consumer"
	"hookify/internal/kafka/publisher"
	"hookify/internal/services/hookify"
	"hookify/internal/storage/postgres"

	grpcapp "hookify/internal/app/grpc"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpcapp.App
	EventQueue *consumer.Consumer
}

func New(log *slog.Logger, grpcPort int) *App {
	storage, err := postgres.New()
	if err != nil {
		log.Error("failed to create postgres storage", "error", err)
		return nil
	}

	publisher := publisher.New(log)
	hookifyService := hookify.New(log, storage, storage, publisher)

	gRPCServer := grpcapp.New(log, hookifyService, grpcPort)
	deliveryService := delivery.New(log, storage)

	return &App{
		log:        log,
		GRPCServer: gRPCServer,
		EventQueue: consumer.New(log, deliveryService),
	}
}
