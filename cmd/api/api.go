package api

import (
	"log/slog"
	"net/http"
	"os"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/starbops/voidrunner/internal/handlers"
	"github.com/starbops/voidrunner/internal/middleware"
	"github.com/starbops/voidrunner/internal/repositories"
	"github.com/starbops/voidrunner/internal/services"
	"github.com/starbops/voidrunner/pkg/auth"
	"github.com/starbops/voidrunner/pkg/config"
)

type APIServer struct {
	addr         string
	config       *config.Config
	taskRepo     repositories.TaskRepository
	userRepo     repositories.UserRepository
	tokenManager *auth.TokenManager
}

func NewAPIServer(addr string, cfg *config.Config, taskRepo repositories.TaskRepository, userRepo repositories.UserRepository) *APIServer {
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiration)
	return &APIServer{
		addr:         addr,
		config:       cfg,
		taskRepo:     taskRepo,
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *APIServer) Run() error {
	mux := http.NewServeMux()

	// Initialize services
	taskService := services.NewTaskService(s.taskRepo)
	userService := services.NewUserService(s.userRepo)
	authService := services.NewAuthService(s.userRepo, s.tokenManager)

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(taskService)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(s.tokenManager)

	// Welcome endpoint (no auth required)
	// Welcome godoc
	//
	//	@Summary		Welcome message
	//	@Description	Returns a welcome message and serves as a health check endpoint
	//	@Tags			System
	//	@Accept			json
	//	@Produce		json
	//	@Success		200	{object}	models.MessageResponse	"Welcome message"
	//	@Router			/welcome [get]
	mux.HandleFunc("GET "+handlers.APIPrefix+"welcome", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Welcome to the VoidRunner API"}`))
	})

	// Swagger documentation endpoint
	mux.HandleFunc("GET /docs/", httpSwagger.WrapHandler)

	// Register authentication routes (no auth required)
	mux.HandleFunc("POST "+handlers.APIPrefix+"register", authHandler.Register)
	mux.HandleFunc("POST "+handlers.APIPrefix+"login", authHandler.Login)
	mux.HandleFunc("POST "+handlers.APIPrefix+"logout", authHandler.Logout)

	// Register protected routes
	taskRouter := taskHandler.RegisterRoutes()
	userRouter := userHandler.RegisterRoutes()
	
	mux.Handle(handlers.APIPrefix+"tasks/", http.StripPrefix(handlers.APIPrefix+"tasks", authMiddleware(taskRouter)))
	mux.Handle(handlers.APIPrefix+"users/", http.StripPrefix(handlers.APIPrefix+"users", authMiddleware(userRouter)))

	server := &http.Server{
		Addr:     s.addr,
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(slog.NewJSONHandler(os.Stdout, nil), slog.LevelError),
	}
	return server.ListenAndServe()
}
