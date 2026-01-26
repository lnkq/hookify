package main

import (
	"log/slog"

	"hookify/internal/app"
)

func main() {
	app := app.New(slog.Default(), 50051)
	app.GRPCServer.MustRun()
}
