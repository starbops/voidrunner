package main

import (
	"log/slog"
	"os"

	"github.com/starbops/voidrunner/cmd/api"
	"github.com/starbops/voidrunner/internal/config"       // Added
	"github.com/starbops/voidrunner/internal/repositories" // Added
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

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("configuration loaded", slog.String("storage_backend", cfg.StorageBackend))

	slog.Info("starting server...",
		slog.String("version", VERSION),
		slog.String("build", BUILD))

	// Initialize repository
	taskRepo, err := repositories.NewTaskRepository(cfg)
	if err != nil {
		slog.Error("failed to initialize task repository", slog.String("error", err.Error()))
		os.Exit(1)
	}

	server := api.NewAPIServer(":8080", taskRepo) // Pass taskRepo
	if err := server.Run(); err != nil {
		slog.Error("failed to start server",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
}
