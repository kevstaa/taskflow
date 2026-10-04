package service

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/kevstaa/taskflow/internal/model"
)

// ---------- Project ----------

type recProjectRepo struct {
	args []string
	err  error
}

func (r *recProjectRepo) Create(_ context.Context, name, description, ownerID string) (*model.Project, error) {
	r.args = []string{name, description, ownerID}
	return &model.Project{}, r.err
}
func (r *recProjectRepo) GetByID(_ context.Context, id string) (*model.Project, error) {
	r.args = []string{id}
	return &model.Project{}, r.err
}
func (r *recProjectRepo) GetByUser(_ context.Context, userID string) ([]model.Project, error) {
	r.args = []string{userID}
	return nil, r.err
}
func (r *recProjectRepo) Update(_ context.Context, id, name, description string) (*model.Project, error) {
	r.args = []string{id, name, description}
	return &model.Project{}, r.err
}
func (r *recProjectRepo) Delete(_ context.Context, id string) error {
	r.args = []string{id}
	return r.err
}

func TestProjectService_ReenviaArgumentosEnOrden(t *testing.T) {
	ctx := context.Background()
	repo := &recProjectRepo{}
	svc := NewProjectService(repo)

	check := func(name string, want ...string) {
		t.Helper()
		if !slices.Equal(repo.args, want) {
			t.Errorf("%s: args = %v, want %v", name, repo.args, want)
		}
	}

	svc.Create(ctx, "nombre", "desc", "owner-1")
	check("Create", "nombre", "desc", "owner-1")

	svc.GetByID(ctx, "p-1")
	check("GetByID", "p-1")

	svc.GetByUser(ctx, "u-1")
	check("GetByUser", "u-1")

	svc.Update(ctx, "p-1", "nombre", "desc")
	check("Update", "p-1", "nombre", "desc")

	svc.Delete(ctx, "p-1")
	check("Delete", "p-1")
}

func TestProjectService_PropagaErrores(t *testing.T) {
	ctx := context.Background()
	svc := NewProjectService(&recProjectRepo{err: errBoom})

	calls := map[string]func() error{
		"Create":    func() error { _, e := svc.Create(ctx, "a", "b", "c"); return e },
		"GetByID":   func() error { _, e := svc.GetByID(ctx, "a"); return e },
		"GetByUser": func() error { _, e := svc.GetByUser(ctx, "a"); return e },
		"Update":    func() error { _, e := svc.Update(ctx, "a", "b", "c"); return e },
		"Delete":    func() error { return svc.Delete(ctx, "a") },
	}
	for name, call := range calls {
		if err := call(); !errors.Is(err, errBoom) {
			t.Errorf("%s: err = %v, want errBoom", name, err)
		}
	}
}

// ---------- ProjectMember ----------

type recMemberRepo struct {
	args     []string
	isMember bool
	err      error
}

func (r *recMemberRepo) AddMember(_ context.Context, projectID, userID, role string) (*model.ProjectMember, error) {
	r.args = []string{projectID, userID, role}
	return &model.ProjectMember{}, r.err
}
func (r *recMemberRepo) GetProjectMembers(_ context.Context, projectID string) ([]model.ProjectMember, error) {
	r.args = []string{projectID}
	return nil, r.err
}
func (r *recMemberRepo) IsMember(_ context.Context, projectID, userID string) (bool, error) {
	r.args = []string{projectID, userID}
	return r.isMember, r.err
}
func (r *recMemberRepo) RemoveMember(_ context.Context, projectID, userID string) error {
	r.args = []string{projectID, userID}
	return r.err
}

func TestProjectMemberService_ReenviaArgumentosEnOrden(t *testing.T) {
	ctx := context.Background()
	repo := &recMemberRepo{}
	svc := NewProjectMemberService(repo)

	svc.AddMember(ctx, "p-1", "u-1", "admin")
	if want := []string{"p-1", "u-1", "admin"}; !slices.Equal(repo.args, want) {
		t.Errorf("AddMember args = %v, want %v", repo.args, want)
	}

	svc.RemoveMember(ctx, "p-1", "u-1")
	if want := []string{"p-1", "u-1"}; !slices.Equal(repo.args, want) {
		t.Errorf("RemoveMember args = %v, want %v", repo.args, want)
	}
}

func TestProjectMemberService_IsMember(t *testing.T) {
	ctx := context.Background()

	for _, want := range []bool{true, false} {
		repo := &recMemberRepo{isMember: want}
		got, err := NewProjectMemberService(repo).IsMember(ctx, "p-1", "u-1")
		if err != nil || got != want {
			t.Errorf("IsMember = %v, %v; want %v, nil", got, err, want)
		}
	}

	repo := &recMemberRepo{err: errBoom}
	if _, err := NewProjectMemberService(repo).IsMember(ctx, "p", "u"); !errors.Is(err, errBoom) {
		t.Errorf("err = %v, want errBoom", err)
	}
}

func TestProjectMemberService_PropagaErrores(t *testing.T) {
	ctx := context.Background()
	svc := NewProjectMemberService(&recMemberRepo{err: errBoom})

	if _, err := svc.AddMember(ctx, "p", "u", "r"); !errors.Is(err, errBoom) {
		t.Errorf("AddMember err = %v", err)
	}
	if _, err := svc.GetProjectMembers(ctx, "p"); !errors.Is(err, errBoom) {
		t.Errorf("GetProjectMembers err = %v", err)
	}
	if err := svc.RemoveMember(ctx, "p", "u"); !errors.Is(err, errBoom) {
		t.Errorf("RemoveMember err = %v", err)
	}
}

// ---------- Task ----------

type recTaskRepo struct {
	args []string
	err  error
}

func (r *recTaskRepo) Create(_ context.Context, title, description, status, projectID, assigneeID string) (*model.Task, error) {
	r.args = []string{title, description, status, projectID, assigneeID}
	return &model.Task{}, r.err
}
func (r *recTaskRepo) GetByID(_ context.Context, id string) (*model.Task, error) {
	r.args = []string{id}
	return &model.Task{}, r.err
}
func (r *recTaskRepo) GetByProject(_ context.Context, projectID string) ([]model.Task, error) {
	r.args = []string{projectID}
	return nil, r.err
}
func (r *recTaskRepo) Update(_ context.Context, id, title, description, status, assigneeID string) (*model.Task, error) {
	r.args = []string{id, title, description, status, assigneeID}
	return &model.Task{}, r.err
}
func (r *recTaskRepo) Delete(_ context.Context, id string) error {
	r.args = []string{id}
	return r.err
}

func TestTaskService_ReenviaArgumentosEnOrden(t *testing.T) {
	ctx := context.Background()
	repo := &recTaskRepo{}
	svc := NewTaskService(repo)

	svc.Create(ctx, "titulo", "desc", "todo", "proj-1", "user-1")
	if want := []string{"titulo", "desc", "todo", "proj-1", "user-1"}; !slices.Equal(repo.args, want) {
		t.Errorf("Create args = %v, want %v", repo.args, want)
	}

	svc.Update(ctx, "t-1", "titulo", "desc", "done", "user-1")
	if want := []string{"t-1", "titulo", "desc", "done", "user-1"}; !slices.Equal(repo.args, want) {
		t.Errorf("Update args = %v, want %v", repo.args, want)
	}

	svc.GetByID(ctx, "t-1")
	svc.GetByProject(ctx, "proj-1")
	if want := []string{"proj-1"}; !slices.Equal(repo.args, want) {
		t.Errorf("GetByProject args = %v, want %v", repo.args, want)
	}
}

func TestTaskService_PropagaErrores(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(&recTaskRepo{err: errBoom})

	calls := map[string]func() error{
		"Create":       func() error { _, e := svc.Create(ctx, "a", "b", "c", "d", "e"); return e },
		"GetByID":      func() error { _, e := svc.GetByID(ctx, "a"); return e },
		"GetByProject": func() error { _, e := svc.GetByProject(ctx, "a"); return e },
		"Update":       func() error { _, e := svc.Update(ctx, "a", "b", "c", "d", "e"); return e },
		"Delete":       func() error { return svc.Delete(ctx, "a") },
	}
	for name, call := range calls {
		if err := call(); !errors.Is(err, errBoom) {
			t.Errorf("%s: err = %v, want errBoom", name, err)
		}
	}
}
