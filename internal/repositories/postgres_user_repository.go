package repositories

import (
	"database/sql"
	"fmt"

	"github.com/starbops/voidrunner/internal/models"
	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	*sql.DB
}

func NewPostgresUserRepository(dataSourceName string) (UserRepository, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresUserRepository{DB: db}, nil
}

func (pur *PostgresUserRepository) GetUsers() ([]*models.User, error) {
	rows, err := pur.Query("SELECT id, username, email, password_hash, first_name, last_name, created_at, updated_at FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (pur *PostgresUserRepository) GetUser(id int) (*models.User, error) {
	var user models.User
	err := pur.QueryRow("SELECT id, username, email, password_hash, first_name, last_name, created_at, updated_at FROM users WHERE id = $1", id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user with id %d: %w", id, err)
	}
	return &user, nil
}

func (pur *PostgresUserRepository) GetByUsernameOrEmail(username, email string) (*models.User, error) {
	var user models.User
	err := pur.QueryRow("SELECT id, username, email, password_hash, first_name, last_name, created_at, updated_at FROM users WHERE username = $1 OR email = $2", username, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username or email: %w", err)
	}
	return &user, nil
}

func (pur *PostgresUserRepository) Create(user *models.User) (*models.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	err := pur.QueryRow(
		"INSERT INTO users (username, email, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at",
		user.Username, user.Email, user.PasswordHash, user.FirstName, user.LastName,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (pur *PostgresUserRepository) CreateUser(user *models.User) (*models.User, error) {
	return pur.Create(user)
}

func (pur *PostgresUserRepository) UpdateUser(id int, user *models.User) (*models.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	err := pur.QueryRow("UPDATE users SET username = $1, email = $2, first_name = $3, last_name = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5 RETURNING updated_at",
		user.Username, user.Email, user.FirstName, user.LastName, id).Scan(&user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update user with id %d: %w", id, err)
	}

	user.ID = id
	return user, nil
}

func (pur *PostgresUserRepository) DeleteUser(id int) error {
	result, err := pur.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete user with id %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for user with id %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %d", id)
	}

	return nil
}