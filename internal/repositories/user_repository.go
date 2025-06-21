package repositories

import (
	"github.com/starbops/voidrunner/internal/models"
)

type UserRepository interface {
	GetUsers() ([]*models.User, error)
	GetUser(id int) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	UpdateUser(id int, user *models.User) (*models.User, error)
	DeleteUser(id int) error
}