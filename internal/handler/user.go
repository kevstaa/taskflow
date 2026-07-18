package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kevstaa/taskflow/internal/middleware"
	"github.com/kevstaa/taskflow/internal/model"
)

type UserServiceInterface interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
	Update(ctx context.Context, id, username, email string) (*model.User, error)
}

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(middleware.UserIDKey).(string)

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(middleware.UserIDKey).(string)

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error": "invalid JSON"}`)
		return
	}

	user, err := h.service.Update(r.Context(), id, req.Username, req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"error": "could not update user"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
