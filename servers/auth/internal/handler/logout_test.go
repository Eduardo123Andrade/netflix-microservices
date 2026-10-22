package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeRefreshRevoker substitui a revogação: devolve o erro configurado e
// registra o token recebido.
type fakeRefreshRevoker struct {
	err error

	called bool
	token  string
}

func (f *fakeRefreshRevoker) Revoke(ctx context.Context, refreshToken string) error {
	f.called = true
	f.token = refreshToken
	return f.err
}

func doLogout(t *testing.T, rv refreshRevoker, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", strings.NewReader(body))
	rec := httptest.NewRecorder()

	Logout(rv).ServeHTTP(rec, req)

	return rec
}

func TestLogoutHandler(t *testing.T) {
	t.Parallel()

	t.Run("No content", func(t *testing.T) {
		t.Parallel()

		rv := &fakeRefreshRevoker{}
		rec := doLogout(t, rv, validRefreshBody)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if rec.Body.Len() != 0 {
			t.Errorf("body = %q, want empty", rec.Body.String())
		}
		if rv.token != "opaque-refresh-token" {
			t.Errorf("Revoke() token = %q, want %q", rv.token, "opaque-refresh-token")
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			body string
		}{
			{name: "empty body", body: ""},
			{name: "malformed JSON", body: `{"refresh_token": `},
			{name: "unknown field", body: `{"refresh_token": "x", "all_devices": true}`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				rv := &fakeRefreshRevoker{}
				rec := doLogout(t, rv, tt.body)

				if rec.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
				}
				if rv.called {
					t.Error("Revoke() called, want it skipped on invalid body")
				}
			})
		}
	})

	t.Run("Error revoking", func(t *testing.T) {
		t.Parallel()

		// Falha de banco não pode virar 204: o cliente acharia que a sessão caiu.
		rec := doLogout(t, &fakeRefreshRevoker{err: errors.New("pq: connection refused to 10.0.0.5")}, validRefreshBody)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if got := decodeError(t, rec); got != "internal error" {
			t.Errorf("error = %q, want %q", got, "internal error")
		}
	})
}
