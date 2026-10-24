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

const (
	fakeUserID       = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
	fakeToken        = "header.payload.signature"
	fakeRefreshToken = "opaque-refresh-token"
)

// fakeRefreshIssuer substitui a emissão de refresh token: devolve um token
// fixo ou o erro configurado e registra para qual usuário foi pedido.
type fakeRefreshIssuer struct {
	err error

	called bool
	userID string
}

func (f *fakeRefreshIssuer) Issue(ctx context.Context, userID string) (string, error) {
	f.called = true
	f.userID = userID
	if f.err != nil {
		return "", f.err
	}
	return fakeRefreshToken, nil
}

// fakeCreateAuth substitui o usecase: devolve o userID e o erro configurados
// e registra o que recebeu.
type fakeCreateAuth struct {
	userID string
	err    error

	called bool
	data   usecase.AuthData
}

func (f *fakeCreateAuth) Execute(ctx context.Context, data usecase.AuthData) (string, error) {
	f.called = true
	f.data = data
	if f.err != nil {
		return "", f.err
	}
	return f.userID, nil
}

// fakeTokenService substitui o token service: devolve um token fixo ou o erro
// configurado e registra para qual usuário foi pedido.
type fakeTokenService struct {
	err error

	called bool
	userID string
}

func (f *fakeTokenService) Generate(userID string) (string, error) {
	f.called = true
	f.userID = userID
	if f.err != nil {
		return "", f.err
	}
	return fakeToken, nil
}

func doCreateAuth(t *testing.T, uc createAuthExecutor, body string) *httptest.ResponseRecorder {
	t.Helper()

	return doCreateAuthWithTokens(t, uc, &fakeTokenService{}, body)
}

func doCreateAuthWithTokens(t *testing.T, uc createAuthExecutor, ts tokenService, body string) *httptest.ResponseRecorder {
	t.Helper()

	return doCreateAuthWithRefresh(t, uc, ts, &fakeRefreshIssuer{}, body)
}

func doCreateAuthWithRefresh(t *testing.T, uc createAuthExecutor, ts tokenService, rt refreshIssuer, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/auth", strings.NewReader(body))
	rec := httptest.NewRecorder()

	CreateAuth(uc, ts, rt).ServeHTTP(rec, req)

	return rec
}

func decodeToken(t *testing.T, rec *httptest.ResponseRecorder) tokenResponse {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
	// Tokens não podem ficar em cache de proxy ou navegador (RFC 6749 §5.1).
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-store")
	}

	// Decodifica num map para conferir também as chaves que não deveriam existir.
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decoding token body %q: %v", rec.Body.String(), err)
	}
	for _, key := range []string{"access_token", "token_type", "refresh_token"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("body = %s, want a %s key", rec.Body.String(), key)
		}
	}

	var resp tokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding token body %q: %v", rec.Body.String(), err)
	}

	return resp
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

	t.Run("Created with token", func(t *testing.T) {
		t.Parallel()

		uc := &fakeCreateAuth{userID: fakeUserID}
		ts := &fakeTokenService{}
		rt := &fakeRefreshIssuer{}
		rec := doCreateAuthWithRefresh(t, uc, ts, rt, validBody)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
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
		// O cliente usa o token_type para montar o header "Authorization: Bearer <token>".
		if got.TokenType != "Bearer" {
			t.Errorf("token_type = %q, want %q", got.TokenType, "Bearer")
		}

		want := usecase.AuthData{
			Name:     "Eduardo",
			Email:    "teste@teste.com",
			Password: "Senha@123",
		}
		if uc.data != want {
			t.Errorf("Execute() data = %+v, want %+v", uc.data, want)
		}
		if ts.userID != fakeUserID {
			t.Errorf("Generate() userID = %q, want %q (the ID returned by the use case)", ts.userID, fakeUserID)
		}
	})

	t.Run("Created without token when generating fails", func(t *testing.T) {
		t.Parallel()

		uc := &fakeCreateAuth{userID: fakeUserID}
		ts := &fakeTokenService{err: errors.New("sign token: key unavailable")}
		rt := &fakeRefreshIssuer{}
		rec := doCreateAuthWithRefresh(t, uc, ts, rt, validBody)

		// A conta foi criada: falhar o token não pode virar erro, o cliente faz login.
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		got := decodeToken(t, rec)
		if got.AccessToken != "" || got.RefreshToken != "" {
			t.Errorf("tokens = (%q, %q), want both empty", got.AccessToken, got.RefreshToken)
		}
		// Sem access token, abrir uma família de refresh só deixaria lixo no banco.
		if rt.called {
			t.Error("Issue() called, want no refresh token when the access token failed")
		}
		if strings.Contains(rec.Body.String(), "key unavailable") {
			t.Errorf("body = %s, want internal details hidden from the client", rec.Body.String())
		}
	})

	t.Run("Created without tokens when issuing refresh fails", func(t *testing.T) {
		t.Parallel()

		rt := &fakeRefreshIssuer{err: errors.New("insert refresh token: connection refused")}
		rec := doCreateAuthWithRefresh(t, &fakeCreateAuth{userID: fakeUserID}, &fakeTokenService{}, rt, validBody)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		// Tudo ou nada: um access token sozinho venceria em minutos sem como renovar.
		got := decodeToken(t, rec)
		if got.AccessToken != "" || got.RefreshToken != "" {
			t.Errorf("tokens = (%q, %q), want both empty", got.AccessToken, got.RefreshToken)
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
				ts := &fakeTokenService{}
				rec := doCreateAuthWithTokens(t, uc, ts, tt.body)

				if ts.called {
					t.Error("Generate() called, want no token on invalid body")
				}
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

				ts := &fakeTokenService{}
				rec := doCreateAuthWithTokens(t, &fakeCreateAuth{err: tt.err}, ts, validBody)

				if ts.called {
					t.Error("Generate() called, want no token when the use case fails")
				}
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
