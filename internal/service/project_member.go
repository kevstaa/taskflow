package service

import (
	"context"

	"github.com/kevstaa/taskflow/internal/model"
)

type ProjectMemberRepo interface {
	AddMember(ctx context.Context, projectID, userID, role string) (*model.ProjectMember, error)
	GetProjectMembers(ctx context.Context, projectID string) ([]model.ProjectMember, error)
	IsMember(ctx context.Context, projectID, userID string) (bool, error)
	RemoveMember(ctx context.Context, projectID, userID string) error
}

type ProjectMemberService struct {
	projectMemberRepo ProjectMemberRepo
}

func NewProjectMemberService(ProjectMemberRepo ProjectMemberRepo) *ProjectMemberService {
	return &ProjectMemberService{projectMemberRepo: ProjectMemberRepo}
}

func (s *ProjectMemberService) AddMember(ctx context.Context, projectID, userID, role string) (*model.ProjectMember, error) {
	return s.projectMemberRepo.AddMember(ctx, projectID, userID, role)
}

func (s *ProjectMemberService) GetProjectMembers(ctx context.Context, projectID string) ([]model.ProjectMember, error) {
	return s.projectMemberRepo.GetProjectMembers(ctx, projectID)
}

func (s *ProjectMemberService) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
	return s.projectMemberRepo.IsMember(ctx, projectID, userID)
}

func (s *ProjectMemberService) RemoveMember(ctx context.Context, projectID, userID string) error {
	return s.projectMemberRepo.RemoveMember(ctx, projectID, userID)
}
