package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kevstaa/taskflow/db"
	"github.com/kevstaa/taskflow/internal/model"
)

type ProjectMemberRepository struct {
	queries *db.Queries
}

func NewProjectMemberRepository(queries *db.Queries) *ProjectMemberRepository {
	return &ProjectMemberRepository{queries: queries}
}

func (r *ProjectMemberRepository) AddMember(ctx context.Context, projectID, userID, role string) (*model.ProjectMember, error) {
	var projectUUID pgtype.UUID
	projectUUID.Scan(projectID)
	var userUUID pgtype.UUID
	userUUID.Scan(userID)

	dbProjectMember, err := r.queries.AddMember(ctx, db.AddMemberParams{
		ProjectID: projectUUID,
		UserID:    userUUID,
		Role:      role,
	})
	if err != nil {
		return nil, err
	}
	return toModelProjectMember(dbProjectMember), nil
}

func (r *ProjectMemberRepository) GetProjectMembers(ctx context.Context, projectID string) ([]model.ProjectMember, error) {
	var uuid pgtype.UUID
	uuid.Scan(projectID)

	dbProjectMembers, err := r.queries.GetProjectMembers(ctx, uuid)
	if err != nil {
		return nil, err
	}

	members := make([]model.ProjectMember, len(dbProjectMembers))
	for i, m := range dbProjectMembers {
		members[i] = toModelProjectMemberFromRow(m)
	}
	return members, nil

}

func (r *ProjectMemberRepository) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
	var projectUUID pgtype.UUID
	projectUUID.Scan(projectID)
	var userUUID pgtype.UUID
	userUUID.Scan(userID)

	return r.queries.IsMember(ctx, db.IsMemberParams{
		ProjectID: projectUUID,
		UserID:    userUUID,
	})
}

func (r *ProjectMemberRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	var projectUUID pgtype.UUID
	projectUUID.Scan(projectID)
	var userUUID pgtype.UUID
	userUUID.Scan(userID)

	return r.queries.RemoveMember(ctx, db.RemoveMemberParams{
		ProjectID: projectUUID,
		UserID:    userUUID,
	})
}

func toModelProjectMember(pm db.ProjectMember) *model.ProjectMember {
	return &model.ProjectMember{
		ProjectID: pm.ProjectID.String(),
		UserID:    pm.UserID.String(),
		Role:      pm.Role,
	}
}

func toModelProjectMemberFromRow(row db.GetProjectMembersRow) model.ProjectMember {
	return model.ProjectMember{
		UserID:   row.ID.String(),
		Role:     row.Role,
		JoinedAt: row.JoinedAt.Time,
	}
}
