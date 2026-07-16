package service

import (
	"context"

	"github.com/kevstaa/taskflow/internal/model"
)

type TaskRepo interface {
	Create(ctx context.Context, title, description, status, projectID, assigneeID string) (*model.Task, error)
	GetByID(ctx context.Context, id string) (*model.Task, error)
	GetByProject(ctx context.Context, projectID string) ([]model.Task, error)
	Update(ctx context.Context, id, title, description, status, assigneeID string) (*model.Task, error)
	Delete(ctx context.Context, id string) error
}

type TaskService struct {
	taskRepo TaskRepo
}

func NewTaskService(TaskRepo TaskRepo) *TaskService {
	return &TaskService{taskRepo: TaskRepo}
}

func (s *TaskService) Create(ctx context.Context, title, description, status, projectID, assigneeID string) (*model.Task, error) {
	return s.taskRepo.Create(ctx, title, description, status, projectID, assigneeID)
}

func (s *TaskService) GetByID(ctx context.Context, id string) (*model.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}

func (s *TaskService) GetByProject(ctx context.Context, projectID string) ([]model.Task, error) {
	return s.taskRepo.GetByProject(ctx, projectID)
}

func (s *TaskService) Update(ctx context.Context, id, title, description, status, assigneeID string) (*model.Task, error) {
	return s.taskRepo.Update(ctx, id, title, description, status, assigneeID)
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	return s.taskRepo.Delete(ctx, id)
}
