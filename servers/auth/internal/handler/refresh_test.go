package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth/internal/usecase"
)

const rotatedRefreshToken = "rotated-refresh-token"

// fakeRefreshRotator substitui a rotação: devolve o userID e o token novo ou
// o erro configurado e registra o token recebido.
type fakeRefreshRotator struct {
	err error

	called bool
	token  string
}

func (f *fakeRefreshRotator) Rotate(ctx context.Context, refreshToken string) (string, string, error) {
	f.called = true
	f.token = refreshToken
	if f.err != nil {
		return "", "", f.err
	}
	return fakeUserID, rotatedRefreshToken, nil
}

func doRefresh(t *testing.T, rot refreshRotator, ts tokenService, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", strings.NewReader(body))
	rec := httptest.NewRecorder()

	Refresh(rot, ts).ServeHTTP(rec, req)

	return rec
}

const validRefreshBody = `{"refresh_token": "opaque-refresh-token"}`

func TestRefreshHandler(t *testing.T) {
	t.Parallel()

	t.Run("OK with a new pair", func(t *testing.T) {
		t.Parallel()

		rot := &fakeRefreshRotator{}
		ts := &fakeTokenService{}
		rec := doRefresh(t, rot, ts, validRefreshBody)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		got := decodeToken(t, rec)
		if got.AccessToken != fakeToken {
			t.Errorf("access_token = %q, want %q", got.AccessToken, fakeToken)
		}
		if got.TokenType != "Bearer" {
			t.Errorf("token_type = %q, want %q", got.TokenType, "Bearer")
		}
		// O cliente precisa trocar o refresh guardado pelo novo: o antigo foi girado.
		if got.RefreshToken != rotatedRefreshToken {
			t.Errorf("refresh_token = %q, want %q (the rotated one)", got.RefreshToken, rotatedRefreshToken)
		}
		if rot.token != "opaque-refresh-token" {
			t.Errorf("Rotate() token = %q, want %q", rot.token, "opaque-refresh-token")
		}
		if ts.userID != fakeUserID {
			t.Errorf("Generate() userID = %q, want %q (the owner of the refresh token)", ts.userID, fakeUserID)
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
			{name: "wrong field type", body: `{"refresh_token": 123}`},
			{name: "unknown field", body: `{"refresh_token": "x", "user_id": "y"}`},
			{name: "body too large", body: `{"refresh_token": "` + strings.Repeat("a", maxBodyBytes) + `"}`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				rot := &fakeRefreshRotator{}
				rec := doRefresh(t, rot, &fakeTokenService{}, tt.body)

				if rec.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
				}
				if got := decodeError(t, rec); got != "invalid request body" {
					t.Errorf("error = %q, want %q", got, "invalid request body")
				}
				if rot.called {
					t.Error("Rotate() called, want it skipped on invalid body")
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
				name:       "invalid refresh token",
				err:        usecase.ErrInvalidRefreshToken,
				wantStatus: http.StatusUnauthorized,
				wantError:  "invalid refresh token",
			},
			{
				// Reuso fora da janela: para o cliente é só "inválido".
				name:       "reused refresh token",
				err:        fmt.Errorf("rotate: %w: %w", usecase.ErrInvalidRefreshToken, usecase.ErrRefreshTokenReused),
				wantStatus: http.StatusUnauthorized,
				wantError:  "invalid refresh token",
			},
			{
				name:       "database error",
				err:        fmt.Errorf("rotate: %w", errors.New("connection refused")),
				wantStatus: http.StatusInternalServerError,
				wantError:  "internal error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				ts := &fakeTokenService{}
				rec := doRefresh(t, &fakeRefreshRotator{err: tt.err}, ts, validRefreshBody)

				if rec.Code != tt.wantStatus {
					t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
				}
				if got := decodeError(t, rec); got != tt.wantError {
					t.Errorf("error = %q, want %q", got, tt.wantError)
				}
				if ts.called {
					t.Error("Generate() called, want no access token when the rotation fails")
				}
			})
		}
	})

	t.Run("Error generating access token", func(t *testing.T) {
		t.Parallel()

		ts := &fakeTokenService{err: errors.New("sign token: key unavailable")}
		rec := doRefresh(t, &fakeRefreshRotator{}, ts, validRefreshBody)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if strings.Contains(rec.Body.String(), rotatedRefreshToken) {
			t.Errorf("body = %s, want no refresh token in an error response", rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "key unavailable") {
			t.Errorf("body = %s, want internal details hidden from the client", rec.Body.String())
		}
	})
}
