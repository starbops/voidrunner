package services

import (
	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/repositories"
)

type UserServiceInterface interface {
	GetUsers() ([]*models.User, error)
	GetUser(id int) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	UpdateUser(id int, user *models.User) (*models.User, error)
	DeleteUser(id int) error
}

type UserService struct {
	userRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (us *UserService) GetUsers() ([]*models.User, error) {
	return us.userRepository.GetUsers()
}

func (us *UserService) GetUser(id int) (*models.User, error) {
	return us.userRepository.GetUser(id)
}

func (us *UserService) CreateUser(user *models.User) (*models.User, error) {
	if user == nil {
		return nil, nil
	}

	return us.userRepository.CreateUser(user)
}

func (us *UserService) UpdateUser(id int, user *models.User) (*models.User, error) {
	if user == nil {
		return nil, nil
	}

	// Set the ID from the path parameter to ensure consistency
	user.ID = id

	return us.userRepository.UpdateUser(id, user)
}

func (us *UserService) DeleteUser(id int) error {
	if id <= 0 {
		return nil
	}

	return us.userRepository.DeleteUser(id)
}