package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth/internal/entity"
	"auth/internal/usecase"

	"golang.org/x/crypto/bcrypt"
)

// fakeLogin substitui o usecase de login: devolve o userID e o erro
// configurados e registra o que recebeu.
type fakeLogin struct {
	userID string
	err    error

	called bool
	data   usecase.LoginData
}

func (f *fakeLogin) Execute(ctx context.Context, data usecase.LoginData) (string, error) {
	f.called = true
	f.data = data
	if f.err != nil {
		return "", f.err
	}
	return f.userID, nil
}

func doLogin(t *testing.T, uc loginExecutor, ts tokenService, body string) *httptest.ResponseRecorder {
	t.Helper()

	return doLoginWithRefresh(t, uc, ts, &fakeRefreshIssuer{}, body)
}

func doLoginWithRefresh(t *testing.T, uc loginExecutor, ts tokenService, rt refreshIssuer, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	Login(uc, ts, rt).ServeHTTP(rec, req)

	return rec
}

const validLoginBody = `{
	"email": "teste@teste.com",
	"password": "Senha@123"
}`

func TestLoginHandler(t *testing.T) {
	t.Parallel()

	t.Run("OK with token", func(t *testing.T) {
		t.Parallel()

		uc := &fakeLogin{userID: fakeUserID}
		ts := &fakeTokenService{}
		rt := &fakeRefreshIssuer{}
		rec := doLoginWithRefresh(t, uc, ts, rt, validLoginBody)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		got := decodeToken(t, rec)
		if got.AccessToken != fakeToken {
			t.Errorf("access_token = %q, want %q", got.AccessToken, fakeToken)
		}
		if got.RefreshToken != fakeRefreshToken {
			t.Errorf("refresh_token = %q, want %q", got.RefreshToken, fakeRefreshToken)
		}
		if rt.userID != fakeUserID {
			t.Errorf("Issue() userID = %q, want %q", rt.userID, fakeUserID)
		}
		if got.TokenType != "Bearer" {
			t.Errorf("token_type = %q, want %q", got.TokenType, "Bearer")
		}

		want := usecase.LoginData{Email: "teste@teste.com", Password: "Senha@123"}
		if uc.data != want {
			t.Errorf("Execute() data = %+v, want %+v", uc.data, want)
		}
		if ts.userID != fakeUserID {
			t.Errorf("Generate() userID = %q, want %q (the ID returned by the use case)", ts.userID, fakeUserID)
		}
	})

	t.Run("Error generating token", func(t *testing.T) {
		t.Parallel()

		// No cadastro a conta já existe e o token é bônus; no login o token é o
		// resultado, então sem ele a resposta é erro.
		ts := &fakeTokenService{err: errors.New("sign token: key unavailable")}
		rec := doLogin(t, &fakeLogin{userID: fakeUserID}, ts, validLoginBody)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if got := decodeError(t, rec); got != "internal error" {
			t.Errorf("error = %q, want %q", got, "internal error")
		}
	})

	t.Run("Error issuing refresh token", func(t *testing.T) {
		t.Parallel()

		rt := &fakeRefreshIssuer{err: errors.New("insert refresh token: connection refused")}
		rec := doLoginWithRefresh(t, &fakeLogin{userID: fakeUserID}, &fakeTokenService{}, rt, validLoginBody)

		// Sem refresh a sessão morreria junto com o access token: não é um login completo.
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if got := decodeError(t, rec); got != "internal error" {
			t.Errorf("error = %q, want %q", got, "internal error")
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
			// O login não aceita campos do cadastro: o contrato é só email e senha.
			{name: "unknown field", body: `{"email": "teste@teste.com", "password": "x", "name": "Eduardo"}`},
			{name: "body too large", body: `{"email": "` + strings.Repeat("a", maxBodyBytes) + `"}`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				uc := &fakeLogin{}
				ts := &fakeTokenService{}
				rec := doLogin(t, uc, ts, tt.body)

				if rec.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
				}
				if got := decodeError(t, rec); got != "invalid request body" {
					t.Errorf("error = %q, want %q", got, "invalid request body")
				}
				if uc.called {
					t.Error("Execute() called, want it skipped on invalid body")
				}
				if ts.called {
					t.Error("Generate() called, want no token on invalid body")
				}
			})
		}
	})

	t.Run("Use case errors", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name       string
			err        error
			wantStatus int
			wantError  string
		}{
			{
				name:       "invalid email format",
				err:        fmt.Errorf("login: %w", entity.ErrInvalidEmail),
				wantStatus: http.StatusBadRequest,
				wantError:  entity.ErrInvalidEmail.Error(),
			},
			{
				name:       "invalid credentials",
				err:        usecase.ErrInvalidCredentials,
				wantStatus: http.StatusUnauthorized,
				wantError:  "invalid credentials",
			},
			{
				// Hash corrompido: o cliente vê o mesmo 401 de uma senha errada.
				name:       "invalid credentials with a server-side cause",
				err:        fmt.Errorf("login: compare hash: %w: %w", usecase.ErrInvalidCredentials, bcrypt.ErrHashTooShort),
				wantStatus: http.StatusUnauthorized,
				wantError:  "invalid credentials",
			},
			{
				name:       "database error",
				err:        fmt.Errorf("login: find by email: %w", errors.New("connection refused")),
				wantStatus: http.StatusInternalServerError,
				wantError:  "internal error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				ts := &fakeTokenService{}
				rec := doLogin(t, &fakeLogin{err: tt.err}, ts, validLoginBody)

				if rec.Code != tt.wantStatus {
					t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
				}
				if got := decodeError(t, rec); got != tt.wantError {
					t.Errorf("error = %q, want %q", got, tt.wantError)
				}
				if ts.called {
					t.Error("Generate() called, want no token when the use case fails")
				}
			})
		}
	})

	t.Run("Error responses do not leak details", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			err    error
			secret string
		}{
			{
				name:   "corrupted hash",
				err:    fmt.Errorf("login: compare hash: %w: %w", usecase.ErrInvalidCredentials, bcrypt.ErrHashTooShort),
				secret: "bcrypt",
			},
			{
				name:   "database error",
				err:    errors.New("pq: connection refused to 10.0.0.5"),
				secret: "10.0.0.5",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				rec := doLogin(t, &fakeLogin{err: tt.err}, &fakeTokenService{}, validLoginBody)

				if strings.Contains(rec.Body.String(), tt.secret) {
					t.Errorf("body = %s, want internal details hidden from the client", rec.Body.String())
				}
			})
		}
	})
}
