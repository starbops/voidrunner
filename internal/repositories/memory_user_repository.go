package repositories

import (
	"sync"
	"time"

	"github.com/starbops/voidrunner/internal/models"
)

type MemoryUserRepository struct {
	users map[int]*models.User
	mutex sync.Mutex
}

func NewMemoryUserRepository() UserRepository {
	return &MemoryUserRepository{
		users: make(map[int]*models.User),
	}
}

func (mur *MemoryUserRepository) GetUsers() ([]*models.User, error) {
	mur.mutex.Lock()
	defer mur.mutex.Unlock()

	users := make([]*models.User, 0, len(mur.users))
	for _, user := range mur.users {
		users = append(users, user)
	}
	return users, nil
}

func (mur *MemoryUserRepository) GetUser(id int) (*models.User, error) {
	mur.mutex.Lock()
	defer mur.mutex.Unlock()

	user, exists := mur.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (mur *MemoryUserRepository) CreateUser(user *models.User) (*models.User, error) {
	mur.mutex.Lock()
	defer mur.mutex.Unlock()

	if user == nil {
		return nil, nil
	}

	if user.ID == 0 {
		maxID := 0
		for id := range mur.users {
			if id > maxID {
				maxID = id
			}
		}
		user.ID = maxID + 1
	}
	if _, exists := mur.users[user.ID]; exists {
		return nil, nil
	}
	if user.CreatedAt == "" {
		user.CreatedAt = time.Now().Format(time.RFC3339Nano)
	}
	if user.UpdatedAt == "" {
		user.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	}

	mur.users[user.ID] = user
	return user, nil
}

func (mur *MemoryUserRepository) UpdateUser(id int, user *models.User) (*models.User, error) {
	mur.mutex.Lock()
	defer mur.mutex.Unlock()

	if user == nil || user.ID != id {
		return nil, nil
	}

	existing, exists := mur.users[id]
	if !exists {
		return nil, nil
	}

	// Preserve the original CreatedAt and update UpdatedAt
	user.CreatedAt = existing.CreatedAt
	user.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	mur.users[id] = user
	return user, nil
}

func (mur *MemoryUserRepository) DeleteUser(id int) error {
	mur.mutex.Lock()
	defer mur.mutex.Unlock()

	if _, exists := mur.users[id]; !exists {
		return nil
	}

	delete(mur.users, id)
	return nil
}