package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kevstaa/taskflow/internal/model"
)

type fakeAuthService struct {
	registerUser *model.User
	registerErr  error
	loginToken   string
	loginErr     error

	gotUsername, gotEmail, gotPassword string
}

func (f *fakeAuthService) Register(_ context.Context, username, email, password string) (*model.User, error) {
	f.gotUsername, f.gotEmail, f.gotPassword = username, email, password
	return f.registerUser, f.registerErr
}

func (f *fakeAuthService) Login(_ context.Context, email, password string) (string, error) {
	f.gotEmail, f.gotPassword = email, password
	return f.loginToken, f.loginErr
}

func post(h http.HandlerFunc, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestRegister(t *testing.T) {
	validBody := `{"username":"kev","email":"kev@example.com","password":"password123"}`

	t.Run("201 y reenvía los datos al servicio", func(t *testing.T) {
		svc := &fakeAuthService{registerUser: &model.User{ID: "1", Username: "kev", Email: "kev@example.com"}}
		rec := post(NewAuthHandler(svc).Register, validBody)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
		if svc.gotUsername != "kev" || svc.gotEmail != "kev@example.com" || svc.gotPassword != "password123" {
			t.Errorf("args al servicio: %q %q %q", svc.gotUsername, svc.gotEmail, svc.gotPassword)
		}
		if !strings.Contains(rec.Body.String(), "kev@example.com") {
			t.Errorf("el body debería incluir el usuario: %s", rec.Body.String())
		}
	})

	t.Run("la respuesta no expone el hash de la contraseña", func(t *testing.T) {
		svc := &fakeAuthService{registerUser: &model.User{
			ID: "1", Username: "kev", Email: "kev@example.com", PasswordHash: "$2a$10$HASH-SECRETO",
		}}
		rec := post(NewAuthHandler(svc).Register, validBody)

		if strings.Contains(rec.Body.String(), "HASH-SECRETO") {
			t.Errorf("el hash de la contraseña aparece en la respuesta: %s", rec.Body.String())
		}
	})

	t.Run("400 con JSON malformado", func(t *testing.T) {
		svc := &fakeAuthService{}
		for _, body := range []string{`{"username":`, ``, `no es json`} {
			rec := post(NewAuthHandler(svc).Register, body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("body %q: status = %d, want 400", body, rec.Code)
			}
		}
	})

	t.Run("400 cuando el servicio falla (comportamiento actual)", func(t *testing.T) {
		svc := &fakeAuthService{registerErr: errors.New("duplicate key")}
		rec := post(NewAuthHandler(svc).Register, validBody)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})
}

func TestLogin(t *testing.T) {
	validBody := `{"email":"kev@example.com","password":"password123"}`

	t.Run("200 con token en JSON", func(t *testing.T) {
		svc := &fakeAuthService{loginToken: "jwt-token"}
		rec := post(NewAuthHandler(svc).Login, validBody)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("body no es JSON válido: %v", err)
		}
		if resp["token"] != "jwt-token" {
			t.Errorf("token = %q", resp["token"])
		}
		if svc.gotEmail != "kev@example.com" || svc.gotPassword != "password123" {
			t.Errorf("args al servicio: %q %q", svc.gotEmail, svc.gotPassword)
		}
	})

	t.Run("401 con credenciales inválidas", func(t *testing.T) {
		svc := &fakeAuthService{loginErr: errors.New("invalid credentials")}
		rec := post(NewAuthHandler(svc).Login, validBody)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
		if strings.Contains(rec.Body.String(), "token") {
			t.Errorf("no debería devolver token: %s", rec.Body.String())
		}
	})

	t.Run("401 con JSON malformado (comportamiento actual)", func(t *testing.T) {
		rec := post(NewAuthHandler(&fakeAuthService{}).Login, `{"email":`)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})
}
