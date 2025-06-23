package repositories

import (
	"sync"
	"time"

	"github.com/starbops/voidrunner/internal/models"
)

type MemoryTaskRepository struct {
	tasks map[int]*models.Task
	mutex sync.Mutex
}

func NewMemoryTaskRepository() TaskRepository {
	return &MemoryTaskRepository{
		tasks: make(map[int]*models.Task),
	}
}

func (mtr *MemoryTaskRepository) GetTasks() ([]*models.Task, error) {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	tasks := make([]*models.Task, 0, len(mtr.tasks))
	for _, task := range mtr.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (mtr *MemoryTaskRepository) GetTask(id int) (*models.Task, error) {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	task, exists := mtr.tasks[id]
	if !exists {
		return nil, nil
	}
	return task, nil
}

func (mtr *MemoryTaskRepository) CreateTask(task *models.Task) (*models.Task, error) {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	if task == nil {
		return nil, nil
	}

	if task.ID == 0 {
		maxID := 0
		for id := range mtr.tasks {
			if id > maxID {
				maxID = id
			}
		}
		task.ID = maxID + 1
	}
	if _, exists := mtr.tasks[task.ID]; exists {
		return nil, nil
	}
	if task.CreatedAt == "" {
		task.CreatedAt = time.Now().Format(time.RFC3339)
	}
	if task.UpdatedAt == "" {
		task.UpdatedAt = time.Now().Format(time.RFC3339)
	}

	mtr.tasks[task.ID] = task
	return task, nil
}

func (mtr *MemoryTaskRepository) UpdateTask(id int, task *models.Task) (*models.Task, error) {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	if task == nil || task.ID != id {
		return nil, nil
	}

	if _, exists := mtr.tasks[id]; !exists {
		return nil, nil
	}

	mtr.tasks[id] = task
	return task, nil
}

func (mtr *MemoryTaskRepository) DeleteTask(id int) error {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	if _, exists := mtr.tasks[id]; !exists {
		return nil
	}

	delete(mtr.tasks, id)
	return nil
}

func (mtr *MemoryTaskRepository) GetTasksByUserID(userID int) ([]*models.Task, error) {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	tasks := make([]*models.Task, 0)
	for _, task := range mtr.tasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (mtr *MemoryTaskRepository) GetTaskByUserID(id, userID int) (*models.Task, error) {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	task, exists := mtr.tasks[id]
	if !exists || task.UserID != userID {
		return nil, nil
	}
	return task, nil
}

func (mtr *MemoryTaskRepository) UpdateTaskByUserID(id, userID int, task *models.Task) (*models.Task, error) {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	if task == nil || task.ID != id {
		return nil, nil
	}

	existingTask, exists := mtr.tasks[id]
	if !exists || existingTask.UserID != userID {
		return nil, nil
	}

	task.UserID = userID
	task.UpdatedAt = time.Now().Format(time.RFC3339)
	mtr.tasks[id] = task
	return task, nil
}

func (mtr *MemoryTaskRepository) DeleteTaskByUserID(id, userID int) error {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()

	task, exists := mtr.tasks[id]
	if !exists || task.UserID != userID {
		return nil
	}

	delete(mtr.tasks, id)
	return nil
}
