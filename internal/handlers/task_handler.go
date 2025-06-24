package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/starbops/voidrunner/internal/middleware"
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

// getTasks godoc
//
//	@Summary		Get user tasks
//	@Description	Retrieve all tasks for the authenticated user
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		models.Task				"List of user tasks"
//	@Failure		401	{object}	models.ErrorResponse	"Unauthorized or user context not found"
//	@Failure		500	{object}	models.ErrorResponse	"Internal server error"
//	@Router			/tasks [get]
func (th *TaskHandler) getTasks(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering getTasks handler")

	userID, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusUnauthorized)
		return
	}

	tasks, err := th.taskService.GetTasksByUserID(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "Failed to encode tasks", http.StatusInternalServerError)
	}
}

// getTask godoc
//
//	@Summary		Get a specific task
//	@Description	Retrieve a specific task by ID for the authenticated user
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int						true	"Task ID"
//	@Success		200	{object}	models.Task				"Task details"
//	@Failure		400	{object}	models.ErrorResponse	"Invalid task ID"
//	@Failure		401	{object}	models.ErrorResponse	"Unauthorized or user context not found"
//	@Failure		404	{object}	models.ErrorResponse	"Task not found"
//	@Failure		500	{object}	models.ErrorResponse	"Internal server error"
//	@Router			/tasks/{id} [get]
func (th *TaskHandler) getTask(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering getTask handler")

	userID, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := th.taskService.GetTaskByUserID(id, userID)
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

// createTask godoc
//
//	@Summary		Create a new task
//	@Description	Create a new task for the authenticated user
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		models.CreateTaskRequest	true	"Task creation data"
//	@Success		201		{object}	models.Task					"Task created successfully"
//	@Failure		400		{object}	models.ErrorResponse		"Invalid request payload"
//	@Failure		401		{object}	models.ErrorResponse		"Unauthorized or user context not found"
//	@Failure		500		{object}	models.ErrorResponse		"Internal server error"
//	@Router			/tasks [post]
func (th *TaskHandler) createTask(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering createTask handler")

	userID, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusUnauthorized)
		return
	}

	var createReq models.CreateTaskRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Convert CreateTaskRequest to Task
	task := models.Task{
		Name:        createReq.Name,
		Description: createReq.Description,
		Status:      createReq.Status,
	}

	createdTask, err := th.taskService.CreateTaskForUser(&task, userID)
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

// updateTask godoc
//
//	@Summary		Update a task
//	@Description	Update an existing task for the authenticated user
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Task ID"
//	@Param			request	body		models.UpdateTaskRequest	true	"Task update data"
//	@Success		200		{object}	models.Task					"Task updated successfully"
//	@Failure		400		{object}	models.ErrorResponse		"Invalid task ID or request payload"
//	@Failure		401		{object}	models.ErrorResponse		"Unauthorized or user context not found"
//	@Failure		404		{object}	models.ErrorResponse		"Task not found"
//	@Failure		500		{object}	models.ErrorResponse		"Internal server error"
//	@Router			/tasks/{id} [put]
func (th *TaskHandler) updateTask(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering updateTask handler")

	userID, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var updateReq models.UpdateTaskRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get existing task first
	existingTask, err := th.taskService.GetTaskByUserID(id, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve task", http.StatusInternalServerError)
		return
	}
	if existingTask == nil {
		http.NotFound(w, req)
		return
	}

	// Apply updates
	if updateReq.Name != nil {
		existingTask.Name = *updateReq.Name
	}
	if updateReq.Description != nil {
		existingTask.Description = *updateReq.Description
	}
	if updateReq.Status != nil {
		existingTask.Status = *updateReq.Status
	}

	updatedTask, err := th.taskService.UpdateTaskByUserID(id, userID, existingTask)
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

// deleteTask godoc
//
//	@Summary		Delete a task
//	@Description	Delete an existing task for the authenticated user
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Task ID"
//	@Success		204	"Task deleted successfully"
//	@Failure		400	{object}	models.ErrorResponse	"Invalid task ID"
//	@Failure		401	{object}	models.ErrorResponse	"Unauthorized or user context not found"
//	@Failure		500	{object}	models.ErrorResponse	"Internal server error"
//	@Router			/tasks/{id} [delete]
func (th *TaskHandler) deleteTask(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering deleteTask handler")

	userID, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := th.taskService.DeleteTaskByUserID(id, userID); err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
