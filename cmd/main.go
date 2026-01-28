package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"hookify/internal/app"
	"hookify/internal/config"

	"github.com/MatusOllah/slogcolor"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if strings.EqualFold(cfg.Env, "production") || strings.EqualFold(cfg.Env, "prod") {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})))
	} else {
		slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, slogcolor.DefaultOptions)))
	}

	application, err := app.New(slog.Default(), cfg)
	if err != nil {
		slog.Error("failed to create app", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		slog.Error("application stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("application stopped")
}
