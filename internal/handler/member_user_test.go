package handler

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kevstaa/taskflow/internal/model"
)

// ---------- ProjectMember ----------

type fakeMemberService struct {
	args    []string
	calls   int
	member  *model.ProjectMember
	members []model.ProjectMember
	err     error
}

func (f *fakeMemberService) AddMember(_ context.Context, projectID, userID, role string) (*model.ProjectMember, error) {
	f.calls++
	f.args = []string{projectID, userID, role}
	return f.member, f.err
}
func (f *fakeMemberService) GetProjectMembers(_ context.Context, projectID string) ([]model.ProjectMember, error) {
	f.calls++
	f.args = []string{projectID}
	return f.members, f.err
}
func (f *fakeMemberService) IsMember(context.Context, string, string) (bool, error) {
	return false, nil
}
func (f *fakeMemberService) RemoveMember(_ context.Context, projectID, userID string) error {
	f.calls++
	f.args = []string{projectID, userID}
	return f.err
}

func memberRouter(h *ProjectMemberHandler) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/projects", func(r chi.Router) {
		r.Get("/{id}/members", h.GetProjectMembers)
		r.Post("/{id}/members", h.AddMember)
		r.Delete("/{id}/members/{userId}", h.RemoveMember)
	})
	return r
}

func TestMemberHandler_AddMember(t *testing.T) {
	body := `{"user_id":"u-1","role":"admin"}`

	t.Run("201 y reenvía projectID (URL), userID y role (body)", func(t *testing.T) {
		svc := &fakeMemberService{member: &model.ProjectMember{ProjectID: "p-1", UserID: "u-1", Role: "admin"}}
		rec := do(memberRouter(NewProjectMemberHandler(svc)), http.MethodPost, "/api/projects/p-1/members", body, "owner-1")

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
		if want := []string{"p-1", "u-1", "admin"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
		if got := decode[model.ProjectMember](t, rec); got.Role != "admin" {
			t.Errorf("miembro = %+v", got)
		}
	})

	t.Run("400 con JSON malformado y no llama al servicio", func(t *testing.T) {
		svc := &fakeMemberService{}
		rec := do(memberRouter(NewProjectMemberHandler(svc)), http.MethodPost, "/api/projects/p-1/members", `{`, "owner-1")

		if rec.Code != http.StatusBadRequest || svc.calls != 0 {
			t.Errorf("status = %d, calls = %d; want 400 y 0", rec.Code, svc.calls)
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeMemberService{err: errBoom}
		rec := do(memberRouter(NewProjectMemberHandler(svc)), http.MethodPost, "/api/projects/p-1/members", body, "owner-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}

func TestMemberHandler_RemoveMember(t *testing.T) {
	t.Run("204 y toma projectID y userId de la URL", func(t *testing.T) {
		svc := &fakeMemberService{}
		rec := do(memberRouter(NewProjectMemberHandler(svc)), http.MethodDelete, "/api/projects/p-1/members/u-1", "", "owner-1")

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", rec.Code)
		}
		if want := []string{"p-1", "u-1"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeMemberService{err: errBoom}
		rec := do(memberRouter(NewProjectMemberHandler(svc)), http.MethodDelete, "/api/projects/p-1/members/u-1", "", "owner-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}

func TestMemberHandler_GetProjectMembers(t *testing.T) {
	svc := &fakeMemberService{members: []model.ProjectMember{{UserID: "u-1"}, {UserID: "u-2"}}}
	rec := do(memberRouter(NewProjectMemberHandler(svc)), http.MethodGet, "/api/projects/p-1/members", "", "user-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if want := []string{"p-1"}; !slices.Equal(svc.args, want) {
		t.Errorf("args = %v, want %v", svc.args, want)
	}
	if got := decode[[]model.ProjectMember](t, rec); len(got) != 2 {
		t.Errorf("devolvió %d miembros, want 2", len(got))
	}

	svc.err = errBoom
	rec = do(memberRouter(NewProjectMemberHandler(svc)), http.MethodGet, "/api/projects/p-1/members", "", "user-1")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("con error: status = %d, want 500", rec.Code)
	}
}

// ---------- User ----------

type fakeUserService struct {
	args  []string
	calls int
	user  *model.User
	err   error
}

func (f *fakeUserService) GetByID(_ context.Context, id string) (*model.User, error) {
	f.calls++
	f.args = []string{id}
	return f.user, f.err
}
func (f *fakeUserService) Update(_ context.Context, id, username, email string) (*model.User, error) {
	f.calls++
	f.args = []string{id, username, email}
	return f.user, f.err
}

func TestUserHandler_GetMe(t *testing.T) {
	t.Run("200 con el user_id del contexto", func(t *testing.T) {
		svc := &fakeUserService{user: &model.User{ID: "u-1", Username: "kev", Email: "kev@example.com"}}
		rec := do(http.HandlerFunc(NewUserHandler(svc).GetMe), http.MethodGet, "/api/users/me", "", "u-1")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if want := []string{"u-1"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
		if got := decode[model.User](t, rec); got.Email != "kev@example.com" {
			t.Errorf("usuario = %+v", got)
		}
	})

	t.Run("no expone el hash de la contraseña", func(t *testing.T) {
		svc := &fakeUserService{user: &model.User{ID: "u-1", PasswordHash: "$2a$10$HASH-SECRETO"}}
		rec := do(http.HandlerFunc(NewUserHandler(svc).GetMe), http.MethodGet, "/api/users/me", "", "u-1")

		if strings.Contains(rec.Body.String(), "HASH-SECRETO") {
			t.Errorf("el hash aparece en la respuesta: %s", rec.Body.String())
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeUserService{err: errBoom}
		rec := do(http.HandlerFunc(NewUserHandler(svc).GetMe), http.MethodGet, "/api/users/me", "", "u-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}

func TestUserHandler_UpdateMe(t *testing.T) {
	body := `{"username":"nuevo","email":"nuevo@example.com"}`

	t.Run("200 y actualiza al usuario del contexto, no al del body", func(t *testing.T) {
		svc := &fakeUserService{user: &model.User{ID: "u-1", Username: "nuevo"}}
		rec := do(http.HandlerFunc(NewUserHandler(svc).UpdateMe), http.MethodPut, "/api/users/me", body, "u-1")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if want := []string{"u-1", "nuevo", "nuevo@example.com"}; !slices.Equal(svc.args, want) {
			t.Errorf("args = %v, want %v", svc.args, want)
		}
	})

	t.Run("400 con JSON malformado y no llama al servicio", func(t *testing.T) {
		svc := &fakeUserService{}
		rec := do(http.HandlerFunc(NewUserHandler(svc).UpdateMe), http.MethodPut, "/api/users/me", `{`, "u-1")

		if rec.Code != http.StatusBadRequest || svc.calls != 0 {
			t.Errorf("status = %d, calls = %d; want 400 y 0", rec.Code, svc.calls)
		}
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := &fakeUserService{err: errBoom}
		rec := do(http.HandlerFunc(NewUserHandler(svc).UpdateMe), http.MethodPut, "/api/users/me", body, "u-1")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}
