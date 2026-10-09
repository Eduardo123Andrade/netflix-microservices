package handler

import (
	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/usecase"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeCreateAuth substitui o usecase: devolve o erro configurado e
// registra o que recebeu.
type fakeCreateAuth struct {
	err error

	called bool
	data   usecase.AuthData
}

func (f *fakeCreateAuth) Execute(ctx context.Context, data usecase.AuthData) error {
	f.called = true
	f.data = data
	return f.err
}

func doCreateAuth(t *testing.T, uc createAuthExecutor, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/auth", strings.NewReader(body))
	rec := httptest.NewRecorder()

	CreateAuth(uc).ServeHTTP(rec, req)

	return rec
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var resp errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding error body %q: %v", rec.Body.String(), err)
	}

	return resp.Error
}

const validBody = `{
	"name": "Eduardo",
	"email": "teste@teste.com",
	"password": "Senha@123"
}`

func TestCreateAuthHandler(t *testing.T) {
	t.Parallel()

	t.Run("Created", func(t *testing.T) {
		t.Parallel()

		uc := &fakeCreateAuth{}
		rec := doCreateAuth(t, uc, validBody)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		if rec.Body.Len() != 0 {
			t.Errorf("body = %q, want empty", rec.Body.String())
		}

		want := usecase.AuthData{
			Name:     "Eduardo",
			Email:    "teste@teste.com",
			Password: "Senha@123",
		}
		if uc.data != want {
			t.Errorf("Execute() data = %+v, want %+v", uc.data, want)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			body string
		}{
			{name: "empty body", body: ""},
			{name: "malformed JSON", body: `{"email": `},
			{name: "not an object", body: `"teste@teste.com"`},
			{name: "wrong field type", body: `{"email": 123}`},
			{name: "unknown field", body: `{"emial": "teste@teste.com"}`},
			{name: "body too large", body: `{"email": "` + strings.Repeat("a", maxBodyBytes) + `"}`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				uc := &fakeCreateAuth{}
				rec := doCreateAuth(t, uc, tt.body)

				if rec.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
				}
				if got := decodeError(t, rec); got != "invalid request body" {
					t.Errorf("error = %q, want %q", got, "invalid request body")
				}
				if uc.called {
					t.Error("Execute() called, want it skipped on invalid body")
				}
			})
		}
	})

	t.Run("Use case errors", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("pq: connection refused to 10.0.0.5")

		tests := []struct {
			name       string
			err        error
			wantStatus int
			wantError  string // trecho que a mensagem precisa conter
		}{
			{
				name:       "invalid email",
				err:        fmt.Errorf("create auth: %w", entity.ErrInvalidEmail),
				wantStatus: http.StatusBadRequest,
				wantError:  entity.ErrInvalidEmail.Error(),
			},
			{
				name:       "invalid password",
				err:        fmt.Errorf("create auth: %w", entity.ErrInvalidPassword),
				wantStatus: http.StatusBadRequest,
				wantError:  entity.ErrInvalidPassword.Error(),
			},
			{
				name:       "invalid email and password",
				err:        fmt.Errorf("create auth: %w", errors.Join(entity.ErrInvalidEmail, entity.ErrInvalidPassword)),
				wantStatus: http.StatusBadRequest,
				wantError:  entity.ErrInvalidPassword.Error(),
			},
			{
				// O userID vem do serviço users: ID inválido é falha interna.
				name:       "invalid user id from users service",
				err:        fmt.Errorf("create auth: new auth: %w", entity.ErrInvalidUserID),
				wantStatus: http.StatusInternalServerError,
				wantError:  "internal error",
			},
			{
				name:       "email already registered",
				err:        repository.ErrAuthAlreadyExists,
				wantStatus: http.StatusConflict,
				wantError:  "email already registered",
			},
			{
				name:       "already registered caught on save",
				err:        fmt.Errorf("create auth: save: %w", repository.ErrAuthAlreadyExists),
				wantStatus: http.StatusConflict,
				wantError:  "email already registered",
			},
			{
				name:       "internal error",
				err:        fmt.Errorf("create auth: find by email: %w", dbErr),
				wantStatus: http.StatusInternalServerError,
				wantError:  "internal error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				rec := doCreateAuth(t, &fakeCreateAuth{err: tt.err}, validBody)

				if rec.Code != tt.wantStatus {
					t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
				}
				if got := decodeError(t, rec); !strings.Contains(got, tt.wantError) {
					t.Errorf("error = %q, want it to contain %q", got, tt.wantError)
				}
			})
		}
	})

	t.Run("Internal error does not leak details", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("pq: connection refused to 10.0.0.5")
		rec := doCreateAuth(t, &fakeCreateAuth{err: dbErr}, validBody)

		if got := decodeError(t, rec); strings.Contains(got, "10.0.0.5") {
			t.Errorf("error = %q, want internal details hidden from the client", got)
		}
	})
}
