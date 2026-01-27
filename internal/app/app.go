package app

import (
	"context"
	"fmt"
	"log/slog"

	"hookify/internal/delivery"
	"hookify/internal/kafka"
	"hookify/internal/services/hookify"
	"hookify/internal/storage/postgres"

	grpcapp "hookify/internal/app/grpcapp"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpcapp.Server
	consumer   *kafka.Consumer
	producer   *kafka.Producer
}

func New(log *slog.Logger, grpcPort int) (*App, error) {
	storage, err := postgres.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres storage: %w", err)
	}

	brokers := []string{"localhost:9092"}
	topic := "webhook.events.raw"
	groupID := "hookify-consumer-1"

	producer := kafka.NewProducer(log, brokers, topic)
	hookifyService := hookify.New(log, storage, storage, producer)

	gRPCServer := grpcapp.New(log, hookifyService, grpcPort)
	deliveryService := delivery.New(log, storage, storage)

	consumer := kafka.NewConsumer(log, brokers, topic, groupID, deliveryService)

	return &App{
		log:        log,
		grpcServer: gRPCServer,
		consumer:   consumer,
		producer:   producer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() { errCh <- a.grpcServer.Run() }()
	go func() { errCh <- a.consumer.Run(ctx) }()

	select {
	case <-ctx.Done():
		a.Stop()
		return nil
	case err := <-errCh:
		a.Stop()
		return err
	}
}

func (a *App) Stop() {
	if a.grpcServer != nil {
		a.grpcServer.Stop()
	}
	if a.consumer != nil {
		if err := a.consumer.Close(); err != nil {
			a.log.Error("failed to close kafka consumer", "error", err)
		}
	}
	if a.producer != nil {
		if err := a.producer.Close(); err != nil {
			a.log.Error("failed to close kafka producer", "error", err)
		}
	}
}
