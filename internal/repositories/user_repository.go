package repositories

import (
	"errors"
	"github.com/starbops/voidrunner/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	GetUsers() ([]*models.User, error)
	GetUser(id int) (*models.User, error)
	GetByUsernameOrEmail(username, email string) (*models.User, error)
	Create(user *models.User) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	UpdateUser(id int, user *models.User) (*models.User, error)
	DeleteUser(id int) error
}