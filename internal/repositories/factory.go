package repositories

func NewTaskRepository() (TaskRepository, error) {
	return NewMemoryTaskRepository(), nil
}
