package main

import (
	"log/slog"
	"os"

	"github.com/starbops/voidrunner/cmd/api"
)

var (
	VERSION = "v0.0.0"
	BUILD   = "dev"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("starting server...",
		slog.String("version", VERSION),
		slog.String("build", BUILD))

	server := api.NewAPIServer(":8080", nil)
	if err := server.Run(); err != nil {
		slog.Error("failed to start server",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
}
