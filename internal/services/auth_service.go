package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/repositories"
	"github.com/starbops/voidrunner/pkg/auth"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound = errors.New("user not found")
)

type AuthServiceInterface interface {
	Register(req models.RegisterRequest) (*models.User, error)
	Login(req models.LoginRequest) (*models.LoginResponse, error)
	Logout(token string) error
}

type AuthService struct {
	userRepo     repositories.UserRepository
	tokenManager *auth.TokenManager
}

func NewAuthService(userRepo repositories.UserRepository, tokenManager *auth.TokenManager) AuthServiceInterface {
	return &AuthService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.User, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("username, email, and password are required")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	existingUser, err := s.userRepo.GetByUsernameOrEmail(req.Username, req.Email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	createdUser, err := s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	if req.Identifier == "" || req.Password == "" {
		return nil, fmt.Errorf("identifier and password are required")
	}

	req.Identifier = strings.TrimSpace(req.Identifier)

	user, err := s.userRepo.GetByUsernameOrEmail(req.Identifier, req.Identifier)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.tokenManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) Logout(token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	err := s.tokenManager.RevokeToken(token)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}