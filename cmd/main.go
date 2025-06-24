// Package main VoidRunner API
//
// VoidRunner is a Go HTTP API server for multi-user task management with JWT authentication.
//
// The API provides endpoints for user registration, authentication, and task management.
// All task operations require JWT authentication and are user-scoped.
//
//	@title			VoidRunner API
//	@version		1.0
//	@description	A multi-user task management API with JWT authentication
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	VoidRunner API Support
//	@contact.url	https://github.com/starbops/voidrunner
//	@contact.email	support@voidrunner.example.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package main

import (
	"log/slog"
	"os"

	"github.com/starbops/voidrunner/cmd/api"
	_ "github.com/starbops/voidrunner/docs" // Import generated docs
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

	server := api.NewAPIServer(":"+cfg.Port, cfg, taskRepo, userRepo)
	if err := server.Run(); err != nil {
		slog.Error("failed to start server",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
}
