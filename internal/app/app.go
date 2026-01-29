package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"hookify/internal/config"
	"hookify/internal/delivery"
	"hookify/internal/kafka"
	"hookify/internal/services/hookify"
	"hookify/internal/storage/postgres"

	grpcapp "hookify/internal/app/grpcapp"
)

type App struct {
	log             *slog.Logger
	grpcServer      *grpcapp.Server
	consumer        *kafka.Consumer
	producer        *kafka.Producer
	storage         *postgres.Storage
	deliveryService *delivery.Service
}

func New(log *slog.Logger, cfg config.Config) (*App, error) {
	storage, err := postgres.New(cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres storage: %w", err)
	}

	producer := kafka.NewProducer(log, cfg.KafkaBrokers, cfg.KafkaTopic)
	hookifyService := hookify.New(log, storage, storage)

	gRPCServer := grpcapp.New(log, hookifyService, cfg.GRPCPort)
	deliveryService := delivery.New(log, storage, storage, storage, producer)
	consumer := kafka.NewConsumer(log, cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, deliveryService, cfg.ConsumerWorkers)

	return &App{
		log:             log,
		grpcServer:      gRPCServer,
		consumer:        consumer,
		producer:        producer,
		storage:         storage,
		deliveryService: deliveryService,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 3)

	go func() { errCh <- a.grpcServer.Run() }()
	go func() { errCh <- a.consumer.Run(ctx) }()
	go func() {
		a.deliveryService.RunOutboxWorker(ctx, time.Second)
		errCh <- nil
	}()

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
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			a.log.Error("failed to close postgres storage", "error", err)
		}
	}
}
