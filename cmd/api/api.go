package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/starbops/voidrunner/internal/handlers"
	"github.com/starbops/voidrunner/internal/repositories"
	"github.com/starbops/voidrunner/internal/services"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	mux := http.NewServeMux()

	// Initialize repositories
	taskRepo, err := repositories.NewTaskRepository()
	if err != nil {
		slog.Error("failed to initialize task repository",
			slog.String("error", err.Error()))
	}

	// Initialize services
	taskService := services.NewTaskService(taskRepo)

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
