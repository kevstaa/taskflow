package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testKey = "test-secret"

func signToken(t *testing.T, key string, claims jwt.MapClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(key))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{"user_id": "user-123", "exp": time.Now().Add(time.Hour).Unix()}
}

func TestAuthenticate_Rechazos(t *testing.T) {
	expired := validClaims()
	expired["exp"] = time.Now().Add(-time.Hour).Unix()

	noneTok, _ := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims()).
		SignedString(jwt.UnsafeAllowNoneSignatureType)

	tests := []struct {
		name     string
		header   string
		wantBody string
	}{
		{"sin header", "", "token required"},
		{"sin prefijo Bearer", signToken(t, testKey, validClaims()), "token required"},
		{"prefijo en minúsculas", "bearer " + signToken(t, testKey, validClaims()), "token required"},
		{"token basura", "Bearer esto-no-es-un-jwt", "invalid token"},
		{"firmado con otra clave", "Bearer " + signToken(t, "otra-clave", validClaims()), "invalid token"},
		{"expirado", "Bearer " + signToken(t, testKey, expired), "invalid token"},
		{"alg=none", "Bearer " + noneTok, "invalid token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
			h := NewAuthMiddleware(testKey).Authenticate(next)

			req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rec.Code)
			}
			if called {
				t.Error("no debería llegar al handler siguiente")
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want que contenga %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestAuthenticate_TokenValido_PonerUserIDEnContexto(t *testing.T) {
	var gotID any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = r.Context().Value(UserIDKey)
		w.WriteHeader(http.StatusOK)
	})
	h := NewAuthMiddleware(testKey).Authenticate(next)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+signToken(t, testKey, validClaims()))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if gotID != "user-123" {
		t.Errorf("user_id en contexto = %v, want user-123", gotID)
	}
}

// Token con firma válida pero claims inesperados: debe responder 401, nunca provocar panic.
func TestAuthenticate_ClaimUserIDInvalido(t *testing.T) {
	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{"sin user_id", jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}},
		{"user_id numérico", jwt.MapClaims{"user_id": 42, "exp": time.Now().Add(time.Hour).Unix()}},
		{"user_id vacío", jwt.MapClaims{"user_id": "", "exp": time.Now().Add(time.Hour).Unix()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("el middleware provocó panic: %v", r)
				}
			}()

			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Error("no debería llegar al handler siguiente")
			})
			h := NewAuthMiddleware(testKey).Authenticate(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+signToken(t, testKey, tt.claims))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rec.Code)
			}
		})
	}
}
