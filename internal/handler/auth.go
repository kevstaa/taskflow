package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kevstaa/taskflow/internal/model"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, username, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type AuthHandler struct {
	service AuthServiceInterface
}

func NewAuthHandler(service AuthServiceInterface) *AuthHandler {
	return &AuthHandler{service: service}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var reg RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "invalid JSON")
		return
	}

	user, err := h.service.Register(r.Context(), reg.Username, reg.Email, reg.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "invalid credentials")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var l LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "invalid JSON")
		return
	}

	token, err := h.service.Login(r.Context(), l.Email, l.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "invalid credentials")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
