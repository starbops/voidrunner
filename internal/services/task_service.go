package services

import (
	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/repositories"
)

type TaskServiceInterface interface {
	GetTasks() ([]*models.Task, error)
	GetTask(id int) (*models.Task, error)
	CreateTask(task *models.Task) (*models.Task, error)
	UpdateTask(id int, task *models.Task) (*models.Task, error)
	DeleteTask(id int) error
}

type TaskService struct {
	taskRepository repositories.TaskRepository
}

func NewTaskService(taskRepository repositories.TaskRepository) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
	}
}

func (ts *TaskService) GetTasks() ([]*models.Task, error) {
	return ts.taskRepository.GetTasks()
}

func (ts *TaskService) GetTask(id int) (*models.Task, error) {
	return ts.taskRepository.GetTask(id)
}

func (ts *TaskService) CreateTask(task *models.Task) (*models.Task, error) {
	if task == nil {
		return nil, nil
	}

	task.Status = models.TaskStatusPending

	return ts.taskRepository.CreateTask(task)
}

func (ts *TaskService) UpdateTask(id int, task *models.Task) (*models.Task, error) {
	if task == nil || task.ID != id {
		return nil, nil
	}

	return ts.taskRepository.UpdateTask(id, task)
}

func (ts *TaskService) DeleteTask(id int) error {
	if id <= 0 {
		return nil
	}

	return ts.taskRepository.DeleteTask(id)
}
