package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"hookify/internal/app"
)

func main() {
	app := app.New(slog.Default(), 50051)

	go app.GRPCServer.MustRun()
	go app.EventQueue.Run(context.Background())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	app.GRPCServer.Stop()

	slog.Info("application stopped")
}
