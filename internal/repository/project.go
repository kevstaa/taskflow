package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kevstaa/taskflow/db"
	"github.com/kevstaa/taskflow/internal/model"
)

type ProjectRepository struct {
	queries *db.Queries
}

func NewProjectRepository(queries *db.Queries) *ProjectRepository {
	return &ProjectRepository{queries: queries}
}

func (r *ProjectRepository) Create(ctx context.Context, name, description, ownerID string) (*model.Project, error) {
	var desc pgtype.Text
	desc.Scan(description)
	var owneruuid pgtype.UUID
	owneruuid.Scan(ownerID)

	dbProject, err := r.queries.CreateProject(ctx, db.CreateProjectParams{
		Name:        name,
		Description: desc,
		OwnerID:     owneruuid,
	})
	if err != nil {
		return nil, err
	}
	return toModelProject(dbProject), nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*model.Project, error) {
	var uuid pgtype.UUID
	uuid.Scan(id)

	dbProject, err := r.queries.GetProjectByID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return toModelProject(dbProject), nil
}

func (r *ProjectRepository) GetByUser(ctx context.Context, userID string) ([]model.Project, error) {
	var uuid pgtype.UUID
	uuid.Scan(userID)

	dbProjects, err := r.queries.GetProjectsByUser(ctx, uuid)
	if err != nil {
		return nil, err
	}

	projects := make([]model.Project, len(dbProjects))
	for i, p := range dbProjects {
		projects[i] = *toModelProject(p)
	}
	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id, name, description string) (*model.Project, error) {
	var desc pgtype.Text
	desc.Scan(description)
	var uuid pgtype.UUID
	uuid.Scan(id)

	dbProject, err := r.queries.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          uuid,
		Name:        name,
		Description: desc,
	})
	if err != nil {
		return nil, err
	}
	return toModelProject(dbProject), nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	var uuid pgtype.UUID
	uuid.Scan(id)
	return r.queries.DeleteProject(ctx, uuid)
}

func toModelProject(p db.Project) *model.Project {
	return &model.Project{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description.String,
		OwnerID:     p.OwnerID.String(),
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}
