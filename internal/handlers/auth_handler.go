package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/services"
)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout)
	return mux
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Create a new user account with username, email, and password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.RegisterRequest	true	"User registration data"
//	@Success		201		{object}	models.User				"User successfully created"
//	@Failure		400		{object}	models.ErrorResponse	"Invalid request body or validation error"
//	@Failure		409		{object}	models.ErrorResponse	"User already exists"
//	@Router			/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Login godoc
//
//	@Summary		Authenticate user
//	@Description	Authenticate user with username/email and password, returns JWT token
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.LoginRequest		true	"User login credentials"
//	@Success		200		{object}	models.LoginResponse	"Login successful"
//	@Failure		400		{object}	models.ErrorResponse	"Invalid request body"
//	@Failure		401		{object}	models.ErrorResponse	"Invalid credentials"
//	@Failure		500		{object}	models.ErrorResponse	"Internal server error"
//	@Router			/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Invalidate the current JWT token and logout the user
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		204	"Logout successful"
//	@Failure		400	{object}	models.ErrorResponse	"Invalid or missing authorization header"
//	@Failure		500	{object}	models.ErrorResponse	"Internal server error"
//	@Router			/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusBadRequest)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
		return
	}

	token := parts[1]
	err := h.authService.Logout(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}