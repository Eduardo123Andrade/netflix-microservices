//go:build integration

package e2e

import (
	"auth/internal/database"
	"auth/internal/router"
	"auth/internal/testutil"
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

const (
	refreshPath = "/api/auth/refresh"
	logoutPath  = "/api/auth/logout"
)

// refreshTokenOf lê o refresh_token de uma resposta de tokens.
func refreshTokenOf(t *testing.T, body string) string {
	t.Helper()

	var resp struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decoding token body %q: %v", body, err)
	}
	if resp.RefreshToken == "" {
		t.Fatalf("body = %q; want a non-empty refresh_token", body)
	}
	return resp.RefreshToken
}

func TestRefreshAndLogoutEndpoints(t *testing.T) {
	t.Parallel()
	_, cfg := testutil.StartPostgres(t)
	testutil.MigrateUp(t, cfg)

	pool, err := database.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	t.Cleanup(pool.Close)

	users := testutil.DialUsers(t, testutil.StartFakeUsers(t, &testutil.FakeUsersServer{}))
	tokens := newTokenService(t)
	srv := newServer(t, router.Deps{DB: pool, Cost: 4, Users: users, Tokens: tokens, RefreshTTL: e2eRefreshTTL})

	if status, body := postJSON(t, srv.URL+createAuthPath, `{
		"email": "refresh@teste.com",
		"name": "Eduardo",
		"password": "Senha@123"
	}`); status != http.StatusCreated {
		t.Fatalf("signup status = %d; want %d (body %q)", status, http.StatusCreated, body)
	}

	login := func(t *testing.T) string {
		t.Helper()

		status, body := postJSON(t, srv.URL+loginPath, `{"email": "refresh@teste.com", "password": "Senha@123"}`)
		if status != http.StatusOK {
			t.Fatalf("login status = %d; want %d (body %q)", status, http.StatusOK, body)
		}
		return refreshTokenOf(t, body)
	}

	refresh := func(t *testing.T, token string) (int, string) {
		t.Helper()

		return postJSON(t, srv.URL+refreshPath, `{"refresh_token": "`+token+`"}`)
	}

	t.Run("refresh gira o token e devolve um access do usuário", func(t *testing.T) {
		t.Parallel()

		first := login(t)

		status, body := refresh(t, first)
		if status != http.StatusOK {
			t.Fatalf("status = %d; want %d (body %q)", status, http.StatusOK, body)
		}
		if sub := tokenSubject(t, tokens, body); sub != testutil.FakeUserID {
			t.Errorf("token sub = %q; want %q", sub, testutil.FakeUserID)
		}
		second := refreshTokenOf(t, body)
		if second == first {
			t.Error("refresh returned the same refresh token, want a rotated one")
		}

		// O novo também funciona, e o antigo ainda vale dentro da janela de 5 minutos.
		if status, body := refresh(t, second); status != http.StatusOK {
			t.Errorf("refresh with the new token status = %d; want %d (body %q)", status, http.StatusOK, body)
		}
		if status, body := refresh(t, first); status != http.StatusOK {
			t.Errorf("refresh with the old token inside the grace period status = %d; want %d (body %q)", status, http.StatusOK, body)
		}
	})

	t.Run("token desconhecido", func(t *testing.T) {
		t.Parallel()

		status, body := refresh(t, "token-que-nunca-existiu")
		if status != http.StatusUnauthorized {
			t.Errorf("status = %d; want %d", status, http.StatusUnauthorized)
		}
		if got := errorMessage(t, body); got != "invalid refresh token" {
			t.Errorf("error = %q; want %q", got, "invalid refresh token")
		}
	})

	t.Run("logout derruba a família inteira", func(t *testing.T) {
		t.Parallel()

		first := login(t)
		_, body := refresh(t, first)
		second := refreshTokenOf(t, body)

		// Logout com o token antigo: a família toda cai, inclusive o girado.
		if status, body := postJSON(t, srv.URL+logoutPath, `{"refresh_token": "`+first+`"}`); status != http.StatusNoContent {
			t.Fatalf("logout status = %d; want %d (body %q)", status, http.StatusNoContent, body)
		}

		for name, token := range map[string]string{"old": first, "rotated": second} {
			if status, _ := refresh(t, token); status != http.StatusUnauthorized {
				t.Errorf("refresh with the %s token after logout status = %d; want %d", name, status, http.StatusUnauthorized)
			}
		}
	})

	t.Run("logout de uma sessão não derruba outra", func(t *testing.T) {
		t.Parallel()

		phone := login(t)
		laptop := login(t)

		if status, _ := postJSON(t, srv.URL+logoutPath, `{"refresh_token": "`+phone+`"}`); status != http.StatusNoContent {
			t.Fatalf("logout status = %d; want %d", status, http.StatusNoContent)
		}
		if status, body := refresh(t, laptop); status != http.StatusOK {
			t.Errorf("refresh on the other session status = %d; want %d (body %q)", status, http.StatusOK, body)
		}
	})

	t.Run("logout com token desconhecido é idempotente", func(t *testing.T) {
		t.Parallel()

		if status, _ := postJSON(t, srv.URL+logoutPath, `{"refresh_token": "token-que-nunca-existiu"}`); status != http.StatusNoContent {
			t.Errorf("status = %d; want %d", status, http.StatusNoContent)
		}
	})
}
