package service

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kevstaa/taskflow/internal/model"
	"golang.org/x/crypto/bcrypt"
)

var errBoom = errors.New("boom")

// ---- fake en memoria de UserRepo ----

type fakeUserRepo struct {
	byEmail   map[string]*model.User
	createErr error
	getErr    error
	nextID    int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]*model.User{}}
}

func (f *fakeUserRepo) Create(ctx context.Context, username, email, hash string) (*model.User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.nextID++
	u := &model.User{ID: strconv.Itoa(f.nextID), Username: username, Email: email, PasswordHash: hash}
	f.byEmail[email] = u
	return u, nil
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	u, ok := f.byEmail[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	for _, u := range f.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeUserRepo) Update(ctx context.Context, id, username, email string) (*model.User, error) {
	u, err := f.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.Username, u.Email = username, email
	return u, nil
}

// ---- AuthService ----

const testKey = "test-secret"

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("hashea la contraseña y guarda username y email", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := NewAuthService(repo, testKey)

		u, err := svc.Register(ctx, "kev", "kev@example.com", "password123")
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if u.Username != "kev" || u.Email != "kev@example.com" {
			t.Errorf("datos incorrectos: %+v", u)
		}
		if u.PasswordHash == "password123" {
			t.Fatal("la contraseña se guardó en claro")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("password123")); err != nil {
			t.Errorf("el hash no corresponde a la contraseña: %v", err)
		}
	})

	t.Run("propaga el error del repositorio", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.createErr = errBoom
		svc := NewAuthService(repo, testKey)

		if _, err := svc.Register(ctx, "kev", "kev@example.com", "password123"); !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want errBoom", err)
		}
	})

	t.Run("contraseña de más de 72 bytes falla (límite de bcrypt)", func(t *testing.T) {
		svc := NewAuthService(newFakeUserRepo(), testKey)
		long := make([]byte, 100)
		for i := range long {
			long[i] = 'a'
		}
		if _, err := svc.Register(ctx, "kev", "kev@example.com", string(long)); err == nil {
			t.Error("se esperaba error por contraseña demasiado larga")
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	setup := func(t *testing.T) (*AuthService, *fakeUserRepo) {
		t.Helper()
		repo := newFakeUserRepo()
		svc := NewAuthService(repo, testKey)
		if _, err := svc.Register(ctx, "kev", "kev@example.com", "password123"); err != nil {
			t.Fatal(err)
		}
		return svc, repo
	}

	t.Run("devuelve un JWT válido con user_id y exp a 24h", func(t *testing.T) {
		svc, repo := setup(t)

		tokenStr, err := svc.Login(ctx, "kev@example.com", "password123")
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		token, err := jwt.Parse(tokenStr, func(tk *jwt.Token) (interface{}, error) {
			if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
				t.Errorf("método de firma inesperado: %v", tk.Header["alg"])
			}
			return []byte(testKey), nil
		})
		if err != nil || !token.Valid {
			t.Fatalf("token inválido: %v", err)
		}

		claims := token.Claims.(jwt.MapClaims)
		if got, want := claims["user_id"], repo.byEmail["kev@example.com"].ID; got != want {
			t.Errorf("user_id = %v, want %v", got, want)
		}

		exp := time.Unix(int64(claims["exp"].(float64)), 0)
		if d := time.Until(exp); d < 23*time.Hour || d > 25*time.Hour {
			t.Errorf("exp a %v, se esperaban ~24h", d)
		}
	})

	t.Run("el token no valida con otra clave", func(t *testing.T) {
		svc, _ := setup(t)
		tokenStr, _ := svc.Login(ctx, "kev@example.com", "password123")

		_, err := jwt.Parse(tokenStr, func(*jwt.Token) (interface{}, error) {
			return []byte("otra-clave"), nil
		})
		if err == nil {
			t.Error("el token no debería validar con otra clave")
		}
	})

	failures := []struct {
		name            string
		email, password string
		repoErr         error
	}{
		{"contraseña incorrecta", "kev@example.com", "mala", nil},
		{"email inexistente", "nadie@example.com", "password123", nil},
		{"error de BD se enmascara como credenciales inválidas", "kev@example.com", "password123", errBoom},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := setup(t)
			repo.getErr = tt.repoErr

			token, err := svc.Login(ctx, tt.email, tt.password)

			if err == nil || err.Error() != "invalid credentials" {
				t.Errorf("err = %v, want 'invalid credentials'", err)
			}
			if token != "" {
				t.Error("no debería devolver token")
			}
		})
	}
}

// ---- UserService ----

func TestUserService(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	u, _ := repo.Create(ctx, "kev", "kev@example.com", "hash")
	svc := NewUserService(repo)

	t.Run("GetByID", func(t *testing.T) {
		got, err := svc.GetByID(ctx, u.ID)
		if err != nil || got.ID != u.ID {
			t.Errorf("got %+v, err %v", got, err)
		}
	})

	t.Run("Update cambia username y email (orden de argumentos)", func(t *testing.T) {
		got, err := svc.Update(ctx, u.ID, "nuevo", "nuevo@example.com")
		if err != nil {
			t.Fatal(err)
		}
		if got.Username != "nuevo" || got.Email != "nuevo@example.com" {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("propaga errores", func(t *testing.T) {
		repo.getErr = errBoom
		if _, err := svc.GetByID(ctx, "x"); !errors.Is(err, errBoom) {
			t.Errorf("GetByID err = %v", err)
		}
		if _, err := svc.Update(ctx, "x", "a", "b"); !errors.Is(err, errBoom) {
			t.Errorf("Update err = %v", err)
		}
	})
}
