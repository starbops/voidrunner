package repositories

import (
	"github.com/starbops/voidrunner/internal/models"
)

type TaskRepository interface {
	GetTasks() ([]*models.Task, error)
	GetTask(id int) (*models.Task, error)
	CreateTask(task *models.Task) (*models.Task, error)
	UpdateTask(id int, task *models.Task) (*models.Task, error)
	DeleteTask(id int) error
}
