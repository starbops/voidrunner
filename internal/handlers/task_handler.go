package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/services"
)

type TaskHandler struct {
	taskService services.TaskServiceInterface
}

func NewTaskHandler(taskService services.TaskServiceInterface) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (th *TaskHandler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", th.getTasks)
	mux.HandleFunc("GET /{id}/", th.getTask)
	mux.HandleFunc("POST /", th.createTask)
	mux.HandleFunc("PUT /{id}/", th.updateTask)
	mux.HandleFunc("DELETE /{id}/", th.deleteTask)

	return mux
}

func (th *TaskHandler) getTasks(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering getTasks handler")

	tasks, err := th.taskService.GetTasks()
	if err != nil {
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "Failed to encode tasks", http.StatusInternalServerError)
	}
}

func (th *TaskHandler) getTask(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering getTask handler")

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := th.taskService.GetTask(id)
	if err != nil {
		http.Error(w, "Failed to retrieve task", http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
	}
}

func (th *TaskHandler) createTask(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering createTask handler")

	var task models.Task
	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdTask, err := th.taskService.CreateTask(&task)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdTask); err != nil {
		http.Error(w, "Failed to encode created task", http.StatusInternalServerError)
	}
}

func (th *TaskHandler) updateTask(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering updateTask handler")

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedTask, err := th.taskService.UpdateTask(id, &task)
	if err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}
	if updatedTask == nil {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
		http.Error(w, "Failed to encode updated task", http.StatusInternalServerError)
	}
}

func (th *TaskHandler) deleteTask(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering deleteTask handler")

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := th.taskService.DeleteTask(id); err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
