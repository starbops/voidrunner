package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/starbops/voidrunner/internal/middleware"
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

	mux.HandleFunc("GET /me", uh.getCurrentUser)
	mux.HandleFunc("PUT /me", uh.updateCurrentUser)
	mux.HandleFunc("DELETE /me", uh.deleteCurrentUser)

	return mux
}

// getCurrentUser godoc
//
//	@Summary		Get current user profile
//	@Description	Retrieve the profile information of the currently authenticated user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.User				"Current user profile"
//	@Failure		401	{object}	models.ErrorResponse	"Unauthorized or user context not found"
//	@Failure		500	{object}	models.ErrorResponse	"Internal server error"
//	@Router			/users/me [get]
func (uh *UserHandler) getCurrentUser(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering getCurrentUser handler")

	userID, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusUnauthorized)
		return
	}

	user, err := uh.userService.GetUser(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
	}
}

// updateCurrentUser godoc
//
//	@Summary		Update current user profile
//	@Description	Update the profile information of the currently authenticated user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		models.UpdateUserRequest	true	"User update data"
//	@Success		200		{object}	models.User					"Updated user profile"
//	@Failure		400		{object}	models.ErrorResponse		"Invalid request payload"
//	@Failure		401		{object}	models.ErrorResponse		"Unauthorized or user context not found"
//	@Failure		500		{object}	models.ErrorResponse		"Internal server error"
//	@Router			/users/me [put]
func (uh *UserHandler) updateCurrentUser(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering updateCurrentUser handler")

	userID, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusUnauthorized)
		return
	}

	var updateReq models.UpdateUserRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get existing user first
	existingUser, err := uh.userService.GetUser(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}
	if existingUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Apply updates
	if updateReq.Username != nil {
		existingUser.Username = *updateReq.Username
	}
	if updateReq.Email != nil {
		existingUser.Email = *updateReq.Email
	}
	if updateReq.FirstName != nil {
		existingUser.FirstName = *updateReq.FirstName
	}
	if updateReq.LastName != nil {
		existingUser.LastName = *updateReq.LastName
	}

	updatedUser, err := uh.userService.UpdateUser(userID, existingUser)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		http.Error(w, "Failed to encode updated user", http.StatusInternalServerError)
	}
}

// deleteCurrentUser godoc
//
//	@Summary		Delete current user account
//	@Description	Delete the account of the currently authenticated user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		204	"User account deleted successfully"
//	@Failure		401	{object}	models.ErrorResponse	"Unauthorized or user context not found"
//	@Failure		500	{object}	models.ErrorResponse	"Internal server error"
//	@Router			/users/me [delete]
func (uh *UserHandler) deleteCurrentUser(w http.ResponseWriter, req *http.Request) {
	slog.Debug("entering deleteCurrentUser handler")

	userID, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusUnauthorized)
		return
	}

	if err := uh.userService.DeleteUser(userID); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}