package models

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
)

// Task represents a task in the system
type Task struct {
	ID          int        `json:"id" example:"1" doc:"Unique identifier for the task"`
	Name        string     `json:"name" example:"Complete project" doc:"Name of the task"`
	Description string     `json:"description" example:"Finish the final report for the project" doc:"Detailed description of the task"`
	Status      TaskStatus `json:"status" example:"pending" enums:"pending,in_progress,completed" doc:"Current status of the task"`
	UserID      int        `json:"user_id" example:"1" doc:"ID of the user who owns this task"`
	CreatedAt   string     `json:"created_at" example:"2023-01-01T10:00:00Z" doc:"Timestamp when the task was created"`
	UpdatedAt   string     `json:"updated_at" example:"2023-01-01T12:00:00Z" doc:"Timestamp when the task was last updated"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Name        string     `json:"name" binding:"required" example:"Complete project" doc:"Name of the task"`
	Description string     `json:"description" example:"Finish the final report for the project" doc:"Description of the task"`
	Status      TaskStatus `json:"status" example:"pending" enums:"pending,in_progress,completed" doc:"Initial status of the task"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Name        *string     `json:"name,omitempty" example:"Updated task name" doc:"Updated name of the task"`
	Description *string     `json:"description,omitempty" example:"Updated description" doc:"Updated description of the task"`
	Status      *TaskStatus `json:"status,omitempty" example:"in_progress" enums:"pending,in_progress,completed" doc:"Updated status of the task"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid input" doc:"Error message"`
	Message string `json:"message,omitempty" example:"The provided data is invalid" doc:"Detailed error message"`
}
