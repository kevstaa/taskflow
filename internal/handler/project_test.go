package handler

import (
	"context"
	"net/http"
	"slices"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kevstaa/taskflow/internal/model"
)

type fakeProjectService struct {
	args     []string
	calls    int
	project  *model.Project
	projects []model.Project
	err      error
}

func (f *fakeProjectService) Create(_ context.Context, name, description, ownerID string) (*model.Project, error) {
	f.calls++
	f.args = []string{name, description, ownerID}
	return f.project, f.err
}
func (f *fakeProjectService) GetByID(_ context.Context, id string) (*model.Project, error) {
	f.calls++
	f.args = []string{id}
	return f.project, f.err
}
func (f *fakeProjectService) GetByUser(_ context.Context, userID string) ([]model.Project, error) {
	f.calls++
	f.args = []string{userID}
	return f.projects, f.err
}
func (f *fakeProjectService) Update(_ context.Context, id, name, description string) (*model.Project, error) {
	f.calls++
	f.args = []string{id, name, description}
	return f.project, f.err
}
func (f *fakeProjectService) Delete(_ context.Context, id string) error {
	f.calls++
	f.args = []string{id}
	return f.err
}

func projectRouter(h *ProjectHandler) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/projects", func(r chi.Router) {
		r.Get("/", h.GetByUser)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
	return r
}

func TestProjectHandler_Create(t *testing.T) {
	body := `{"name":"Mi proyecto","description":"desc"}`

	t.Run("201 y usa el user_id del contexto como owner", func(t *testing.T) {
		svc := &fakeProjectService{project: &model.Project{ID: "p-1", Name: "Mi proyecto"}}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodPost, "/api/projects", body, "user-1")

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		if want := []string{"Mi proyecto", "desc", "user-1"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
		if got := decode[model.Project](t, rec); got.ID != "p-1" {
			t.Errorf("proyecto devuelto = %+v", got)
		}
	})

	t.Run("400 con JSON malformado y no llama al servicio", func(t *testing.T) {
		svc := &fakeProjectService{}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodPost, "/api/projects", `{"name":`, "user-1")

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
		if svc.calls != 0 {
			t.Error("no debería llamar al servicio")
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeProjectService{err: errBoom}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodPost, "/api/projects", body, "user-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}

func TestProjectHandler_GetByID(t *testing.T) {
	t.Run("200 con el id de la URL", func(t *testing.T) {
		svc := &fakeProjectService{project: &model.Project{ID: "p-1", Name: "X"}}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodGet, "/api/projects/p-1", "", "user-1")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if want := []string{"p-1"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
		if got := decode[model.Project](t, rec); got.ID != "p-1" {
			t.Errorf("proyecto = %+v", got)
		}
	})

	t.Run("500 si falla el servicio (comportamiento actual, incluso para 'no encontrado')", func(t *testing.T) {
		svc := &fakeProjectService{err: errBoom}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodGet, "/api/projects/p-1", "", "user-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}

func TestProjectHandler_GetByUser(t *testing.T) {
	t.Run("200 con el user_id del contexto", func(t *testing.T) {
		svc := &fakeProjectService{projects: []model.Project{{ID: "p-1"}, {ID: "p-2"}}}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodGet, "/api/projects", "", "user-1")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if want := []string{"user-1"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
		if got := decode[[]model.Project](t, rec); len(got) != 2 {
			t.Errorf("devolvió %d proyectos, want 2", len(got))
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeProjectService{err: errBoom}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodGet, "/api/projects", "", "user-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}

func TestProjectHandler_Update(t *testing.T) {
	body := `{"name":"Nuevo","description":"nueva desc"}`

	t.Run("200 y reenvía id, name, description en orden", func(t *testing.T) {
		svc := &fakeProjectService{project: &model.Project{ID: "p-1", Name: "Nuevo"}}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodPut, "/api/projects/p-1", body, "user-1")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if want := []string{"p-1", "Nuevo", "nueva desc"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
	})

	t.Run("400 con JSON malformado y no llama al servicio", func(t *testing.T) {
		svc := &fakeProjectService{}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodPut, "/api/projects/p-1", `nope`, "user-1")

		if rec.Code != http.StatusBadRequest || svc.calls != 0 {
			t.Errorf("status = %d, calls = %d; want 400 y 0", rec.Code, svc.calls)
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeProjectService{err: errBoom}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodPut, "/api/projects/p-1", body, "user-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}

func TestProjectHandler_Delete(t *testing.T) {
	t.Run("204 sin body y con el id de la URL", func(t *testing.T) {
		svc := &fakeProjectService{}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodDelete, "/api/projects/p-1", "", "user-1")

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Errorf("body = %q, want vacío", rec.Body.String())
		}
		if want := []string{"p-1"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeProjectService{err: errBoom}
		rec := do(projectRouter(NewProjectHandler(svc)), http.MethodDelete, "/api/projects/p-1", "", "user-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}
