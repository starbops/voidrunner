package api

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/starbops/voidrunner/internal/handlers"
	"github.com/starbops/voidrunner/internal/repositories"
	"github.com/starbops/voidrunner/internal/services"
)

type APIServer struct {
	addr     string
	taskRepo repositories.TaskRepository
	userRepo repositories.UserRepository
}

func NewAPIServer(addr string, taskRepo repositories.TaskRepository, userRepo repositories.UserRepository) *APIServer {
	return &APIServer{
		addr:     addr,
		taskRepo: taskRepo,
		userRepo: userRepo,
	}
}

func (s *APIServer) Run() error {
	mux := http.NewServeMux()

	// Initialize services
	taskService := services.NewTaskService(s.taskRepo)
	userService := services.NewUserService(s.userRepo)

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(taskService)
	taskRouter := taskHandler.RegisterRoutes()

	userHandler := handlers.NewUserHandler(userService)
	userRouter := userHandler.RegisterRoutes()

	// Register API routes
	mux.HandleFunc(handlers.APIPrefix, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Welcome to the VoidRunner API"}`))
	})
	mux.Handle(handlers.APIPrefix+"tasks/", http.StripPrefix(handlers.APIPrefix+"tasks", taskRouter))
	mux.Handle(handlers.APIPrefix+"users/", http.StripPrefix(handlers.APIPrefix+"users", userRouter))

	server := &http.Server{
		Addr:     s.addr,
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(slog.NewJSONHandler(os.Stdout, nil), slog.LevelError),
	}
	return server.ListenAndServe()
}
