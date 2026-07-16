package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kevstaa/taskflow/db"
	"github.com/kevstaa/taskflow/internal/model"
)

type TaskRepository struct {
	queries *db.Queries
}

func NewTaskRepository(queries *db.Queries) *TaskRepository {
	return &TaskRepository{queries: queries}
}

func (r *TaskRepository) Create(ctx context.Context, title, description, status, projectID, assigneeID string) (*model.Task, error) {
	var desc pgtype.Text
	desc.Scan(description)
	var projectUUID pgtype.UUID
	projectUUID.Scan(projectID)
	var assigneeUUID pgtype.UUID
	assigneeUUID.Scan(assigneeID)

	dbTask, err := r.queries.CreateTask(ctx, db.CreateTaskParams{
		Title:       title,
		Description: desc,
		Status:      status,
		ProjectID:   projectUUID,
		AssigneeID:  assigneeUUID,
	})
	if err != nil {
		return nil, err
	}
	return toModelTask(dbTask), err
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	var uuid pgtype.UUID
	uuid.Scan(id)

	dbTask, err := r.queries.GetTaskByID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return toModelTask(dbTask), nil
}

func (r *TaskRepository) GetByProject(ctx context.Context, projectID string) ([]model.Task, error) {
	var projectUUID pgtype.UUID
	projectUUID.Scan(projectID)

	dbTasks, err := r.queries.GetTasksByProject(ctx, projectUUID)
	if err != nil {
		return nil, err
	}

	tasks := make([]model.Task, len(dbTasks))
	for i, t := range dbTasks {
		tasks[i] = *toModelTask(t)
	}
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, id, title, description, status, assigneeID string) (*model.Task, error) {
	var uuid pgtype.UUID
	uuid.Scan(id)
	var desc pgtype.Text
	desc.Scan(description)
	var assigneeUUID pgtype.UUID
	assigneeUUID.Scan(assigneeID)

	dbTask, err := r.queries.UpdateTask(ctx, db.UpdateTaskParams{
		ID:          uuid,
		Title:       title,
		Description: desc,
		Status:      status,
		AssigneeID:  assigneeUUID,
	})
	if err != nil {
		return nil, err
	}
	return toModelTask(dbTask), nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	var uuid pgtype.UUID
	uuid.Scan(id)
	return r.queries.DeleteTask(ctx, uuid)
}

func toModelTask(t db.Task) *model.Task {
	var assigneeID *string
	if t.AssigneeID.Valid {
		s := t.AssigneeID.String()
		assigneeID = &s
	}

	return &model.Task{
		Title:       t.Title,
		Description: t.Description.String,
		Status:      t.Status,
		ProjectID:   t.ProjectID.String(),
		AssigneeID:  assigneeID,
	}
}
