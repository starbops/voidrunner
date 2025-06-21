package main

import (
	"log/slog"
	"os"

	"github.com/starbops/voidrunner/cmd/api"
	"github.com/starbops/voidrunner/internal/repositories"
	"github.com/starbops/voidrunner/pkg/config"
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

	slog.Info("loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize repositories
	taskRepo, err := repositories.NewTaskRepository(cfg)
	if err != nil {
		slog.Error("failed to initialize task repository",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	userRepo, err := repositories.NewUserRepository(cfg)
	if err != nil {
		slog.Error("failed to initialize user repository",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("starting server...",
		slog.String("version", VERSION),
		slog.String("build", BUILD))

	server := api.NewAPIServer(":8080", taskRepo, userRepo)
	if err := server.Run(); err != nil {
		slog.Error("failed to start server",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
}
