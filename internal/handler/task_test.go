package handler

import (
	"context"
	"net/http"
	"slices"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kevstaa/taskflow/internal/model"
)

type fakeTaskService struct {
	args  []string
	calls int
	task  *model.Task
	tasks []model.Task
	err   error
}

func (f *fakeTaskService) Create(_ context.Context, title, description, status, projectID, assigneeID string) (*model.Task, error) {
	f.calls++
	f.args = []string{title, description, status, projectID, assigneeID}
	return f.task, f.err
}
func (f *fakeTaskService) GetByID(_ context.Context, id string) (*model.Task, error) {
	f.calls++
	f.args = []string{id}
	return f.task, f.err
}
func (f *fakeTaskService) GetByProject(_ context.Context, projectID string) ([]model.Task, error) {
	f.calls++
	f.args = []string{projectID}
	return f.tasks, f.err
}
func (f *fakeTaskService) Update(_ context.Context, id, title, description, status, assigneeID string) (*model.Task, error) {
	f.calls++
	f.args = []string{id, title, description, status, assigneeID}
	return f.task, f.err
}
func (f *fakeTaskService) Delete(_ context.Context, id string) error {
	f.calls++
	f.args = []string{id}
	return f.err
}

// Mismas rutas que cmd/api/main.go
func taskRouter(h *TaskHandler) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/projects/{id}/tasks", func(r chi.Router) {
		r.Get("/", h.GetByProject)
		r.Post("/", h.Create)
		r.Get("/{tid}", h.GetByID)
		r.Put("/{tid}", h.Update)
		r.Delete("/{tid}", h.Delete)
	})
	return r
}

func TestTaskHandler_Create(t *testing.T) {
	body := `{"title":"T","description":"D","status":"todo","assignee_id":"u-9"}`

	t.Run("201 y toma projectID de la URL", func(t *testing.T) {
		svc := &fakeTaskService{task: &model.Task{ID: "t-1", Title: "T"}}
		rec := do(taskRouter(NewTaskHandler(svc)), http.MethodPost, "/api/projects/proj-1/tasks", body, "user-1")

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
		if want := []string{"T", "D", "todo", "proj-1", "u-9"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
		if got := decode[model.Task](t, rec); got.ID != "t-1" {
			t.Errorf("tarea = %+v", got)
		}
	})

	t.Run("400 con JSON malformado y no llama al servicio", func(t *testing.T) {
		svc := &fakeTaskService{}
		rec := do(taskRouter(NewTaskHandler(svc)), http.MethodPost, "/api/projects/proj-1/tasks", `{`, "user-1")

		if rec.Code != http.StatusBadRequest || svc.calls != 0 {
			t.Errorf("status = %d, calls = %d; want 400 y 0", rec.Code, svc.calls)
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeTaskService{err: errBoom}
		rec := do(taskRouter(NewTaskHandler(svc)), http.MethodPost, "/api/projects/proj-1/tasks", body, "user-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}

func TestTaskHandler_GetByProject(t *testing.T) {
	svc := &fakeTaskService{tasks: []model.Task{{ID: "t-1"}, {ID: "t-2"}}}
	rec := do(taskRouter(NewTaskHandler(svc)), http.MethodGet, "/api/projects/proj-1/tasks", "", "user-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if want := []string{"proj-1"}; !slices.Equal(svc.args, want) {
		t.Errorf("args = %v, want %v", svc.args, want)
	}
	if got := decode[[]model.Task](t, rec); len(got) != 2 {
		t.Errorf("devolvió %d tareas, want 2", len(got))
	}

	svc.err = errBoom
	rec = do(taskRouter(NewTaskHandler(svc)), http.MethodGet, "/api/projects/proj-1/tasks", "", "user-1")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("con error: status = %d, want 500", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// OJO: los tres tests siguientes FALLAN con el código actual.
// GetByID, Update y Delete leen chi.URLParam(r, "id"), que en la ruta
// /api/projects/{id}/tasks/{tid} es el ID del PROYECTO. El de la tarea es "tid".
// ---------------------------------------------------------------------------

func TestTaskHandler_GetByID_UsaElIDDeLaTarea(t *testing.T) {
	svc := &fakeTaskService{task: &model.Task{ID: "task-1"}}
	rec := do(taskRouter(NewTaskHandler(svc)), http.MethodGet, "/api/projects/proj-1/tasks/task-1", "", "user-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if want := []string{"task-1"}; !slices.Equal(svc.args, want) {
		t.Errorf("el servicio recibió %v, want %v (¿se está usando el id del proyecto?)", svc.args, want)
	}
}

func TestTaskHandler_Update_UsaElIDDeLaTarea(t *testing.T) {
	svc := &fakeTaskService{task: &model.Task{ID: "task-1"}}
	body := `{"title":"T","description":"D","status":"done","assignee_id":"u-9"}`
	rec := do(taskRouter(NewTaskHandler(svc)), http.MethodPut, "/api/projects/proj-1/tasks/task-1", body, "user-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if want := []string{"task-1", "T", "D", "done", "u-9"}; !slices.Equal(svc.args, want) {
		t.Errorf("el servicio recibió %v, want %v", svc.args, want)
	}
}

func TestTaskHandler_Delete_UsaElIDDeLaTarea(t *testing.T) {
	svc := &fakeTaskService{}
	rec := do(taskRouter(NewTaskHandler(svc)), http.MethodDelete, "/api/projects/proj-1/tasks/task-1", "", "user-1")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if want := []string{"task-1"}; !slices.Equal(svc.args, want) {
		t.Errorf("el servicio recibió %v, want %v", svc.args, want)
	}
}

func TestTaskHandler_Update_ErroresYJSONInvalido(t *testing.T) {
	svc := &fakeTaskService{}
	rec := do(taskRouter(NewTaskHandler(svc)), http.MethodPut, "/api/projects/proj-1/tasks/task-1", `{`, "user-1")
	if rec.Code != http.StatusBadRequest || svc.calls != 0 {
		t.Errorf("JSON malformado: status = %d, calls = %d; want 400 y 0", rec.Code, svc.calls)
	}

	svc = &fakeTaskService{err: errBoom}
	rec = do(taskRouter(NewTaskHandler(svc)), http.MethodPut, "/api/projects/proj-1/tasks/task-1", `{}`, "user-1")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("error del servicio: status = %d, want 500", rec.Code)
	}
}
