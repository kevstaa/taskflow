package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kevstaa/taskflow/internal/middleware"
)

var errBoom = errors.New("boom")

// do lanza una petición contra h. Si userID != "" lo inyecta en el contexto,
// igual que haría el middleware de autenticación.
func do(h http.Handler, method, path, body, userID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if userID != "" {
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("body no es JSON válido: %v (body: %q)", err, rec.Body.String())
	}
	return v
}
