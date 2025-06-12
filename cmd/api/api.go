package api

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/starbops/voidrunner/internal/handlers"
	"github.com/starbops/voidrunner/internal/repositories" // Already present
	"github.com/starbops/voidrunner/internal/services"
)

type APIServer struct {
	addr     string
	taskRepo repositories.TaskRepository // Changed from db *sql.DB
}

func NewAPIServer(addr string, taskRepo repositories.TaskRepository) *APIServer { // Changed db to taskRepo
	return &APIServer{
		addr:     addr,
		taskRepo: taskRepo, // Initialize the new field
	}
}

func (s *APIServer) Run() error {
	mux := http.NewServeMux()

	// Initialize services
	// taskRepo is now a field of APIServer (s.taskRepo)
	// The local taskRepo initialization is removed from here.
	taskService := services.NewTaskService(s.taskRepo) // Use the passed-in repository

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(taskService)
	taskRouter := taskHandler.RegisterRoutes()

	// Register API routes
	mux.HandleFunc(handlers.APIPrefix, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Welcome to the VoidRunner API"}`))
	})
	mux.Handle(handlers.APIPrefix+"tasks/", http.StripPrefix(handlers.APIPrefix+"tasks", taskRouter))

	server := &http.Server{
		Addr:     s.addr,
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(slog.NewJSONHandler(os.Stdout, nil), slog.LevelError),
	}
	return server.ListenAndServe()
}
