package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kevstaa/taskflow/internal/model"
)

type ProjectMemberServiceInterface interface {
	AddMember(ctx context.Context, projectID, userID, role string) (*model.ProjectMember, error)
	GetProjectMembers(ctx context.Context, projectID string) ([]model.ProjectMember, error)
	IsMember(ctx context.Context, projectID, userID string) (bool, error)
	RemoveMember(ctx context.Context, projectID, userID string) error
}

type ProjectMemberHandler struct {
	service ProjectMemberServiceInterface
}

func NewProjectMemberHandler(service ProjectMemberServiceInterface) *ProjectMemberHandler {
	return &ProjectMemberHandler{service: service}
}

type AddMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (h *ProjectMemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error": "invalid JSON"}`)
		return
	}

	projectMember, err := h.service.AddMember(r.Context(), id, req.UserID, req.Role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"error": "could not add member"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(projectMember)
}

func (h *ProjectMemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")

	err := h.service.RemoveMember(r.Context(), projectID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectMemberHandler) GetProjectMembers(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	users, err := h.service.GetProjectMembers(r.Context(), projectID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
