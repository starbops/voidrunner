package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/services"
)

type UserHandler struct {
	userService services.UserServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (uh *UserHandler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", uh.getUsers)
	mux.HandleFunc("GET /{id}/", uh.getUser)
	mux.HandleFunc("POST /", uh.createUser)
	mux.HandleFunc("PUT /{id}/", uh.updateUser)
	mux.HandleFunc("DELETE /{id}/", uh.deleteUser)

	return mux
}

func (uh *UserHandler) getUsers(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering getUsers handler")

	users, err := uh.userService.GetUsers()
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode users", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) getUser(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering getUser handler")

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := uh.userService.GetUser(id)
	if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) createUser(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering createUser handler")

	var user models.User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdUser, err := uh.userService.CreateUser(&user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdUser); err != nil {
		http.Error(w, "Failed to encode created user", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) updateUser(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering updateUser handler")

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedUser, err := uh.userService.UpdateUser(id, &user)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	if updatedUser == nil {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		http.Error(w, "Failed to encode updated user", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) deleteUser(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering deleteUser handler")

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := uh.userService.DeleteUser(id); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}