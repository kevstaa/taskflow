package service

import (
	"context"

	"github.com/kevstaa/taskflow/internal/model"
)

type ProjectRepo interface {
	Create(ctx context.Context, name, description, ownerID string) (*model.Project, error)
	GetByID(ctx context.Context, id string) (*model.Project, error)
	GetByUser(ctx context.Context, userID string) ([]model.Project, error)
	Update(ctx context.Context, id, name, description string) (*model.Project, error)
	Delete(ctx context.Context, id string) error
}

type ProjectService struct {
	projectRepo ProjectRepo
}

func NewProjectService(ProjectRepo ProjectRepo) *ProjectService {
	return &ProjectService{projectRepo: ProjectRepo}
}

func (s *ProjectService) Create(ctx context.Context, name, description, ownerID string) (*model.Project, error) {
	return s.projectRepo.Create(ctx, name, description, ownerID)
}

func (s *ProjectService) GetByID(ctx context.Context, id string) (*model.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

func (s *ProjectService) GetByUser(ctx context.Context, userID string) ([]model.Project, error) {
	return s.projectRepo.GetByUser(ctx, userID)
}

func (s *ProjectService) Update(ctx context.Context, id, name, description string) (*model.Project, error) {
	return s.projectRepo.Update(ctx, id, name, description)
}

func (s *ProjectService) Delete(ctx context.Context, id string) error {
	return s.projectRepo.Delete(ctx, id)
}
