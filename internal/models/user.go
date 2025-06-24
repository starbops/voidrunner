package models

// User represents a user in the system
type User struct {
	ID           int    `json:"id" example:"1" doc:"Unique identifier for the user"`
	Username     string `json:"username" example:"johndoe" doc:"Username of the user"`
	Email        string `json:"email" example:"john.doe@example.com" doc:"Email address of the user"`
	PasswordHash string `json:"-" doc:"Hashed password (never returned in API responses)"`
	FirstName    string `json:"first_name" example:"John" doc:"First name of the user"`
	LastName     string `json:"last_name" example:"Doe" doc:"Last name of the user"`
	CreatedAt    string `json:"created_at" example:"2023-01-01T10:00:00Z" doc:"Timestamp when the user was created"`
	UpdatedAt    string `json:"updated_at" example:"2023-01-01T12:00:00Z" doc:"Timestamp when the user was last updated"`
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username  string `json:"username" binding:"required" example:"johndoe" doc:"Desired username (must be unique)"`
	Email     string `json:"email" binding:"required" example:"john.doe@example.com" doc:"Email address (must be unique)"`
	Password  string `json:"password" binding:"required" example:"securepassword123" doc:"Password (minimum 6 characters)"`
	FirstName string `json:"first_name,omitempty" example:"John" doc:"First name (optional)"`
	LastName  string `json:"last_name,omitempty" example:"Doe" doc:"Last name (optional)"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required" example:"johndoe" doc:"Username or email address"`
	Password   string `json:"password" binding:"required" example:"securepassword123" doc:"User password"`
}

// LoginResponse represents the response body for successful login
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." doc:"JWT authentication token"`
	User  User   `json:"user" doc:"User information"`
}

// UpdateUserRequest represents the request body for updating user information
type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty" example:"newusername" doc:"Updated username"`
	Email     *string `json:"email,omitempty" example:"newemail@example.com" doc:"Updated email address"`
	FirstName *string `json:"first_name,omitempty" example:"NewFirstName" doc:"Updated first name"`
	LastName  *string `json:"last_name,omitempty" example:"NewLastName" doc:"Updated last name"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully" doc:"Response message"`
}