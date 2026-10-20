//go:build integration

package e2e

import (
	"auth/internal/database"
	"auth/internal/router"
	"auth/internal/testutil"
	"context"
	"net/http"
	"testing"
)

const loginPath = "/api/auth/login"

func TestLoginEndpoint(t *testing.T) {
	t.Parallel()
	_, cfg := testutil.StartPostgres(t)
	testutil.MigrateUp(t, cfg)

	pool, err := database.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	t.Cleanup(pool.Close)

	usersAddr := testutil.StartFakeUsers(t, &testutil.FakeUsersServer{})
	users := testutil.DialUsers(t, usersAddr)
	tokens := newTokenService(t)

	srv := newServer(t, router.Deps{DB: pool, Cost: 4, Users: users, Tokens: tokens})

	// A conta é criada pelo próprio endpoint de cadastro: o login lê o que o
	// cadastro gravou, com o hash bcrypt de verdade.
	if status, body := postJSON(t, srv.URL+createAuthPath, `{
		"email": "login@teste.com",
		"name": "Eduardo",
		"password": "Senha@123"
	}`); status != http.StatusCreated {
		t.Fatalf("signup status = %d; want %d (body %q)", status, http.StatusCreated, body)
	}

	url := srv.URL + loginPath

	t.Run("credenciais certas devolvem token do usuário", func(t *testing.T) {
		t.Parallel()

		status, body := postJSON(t, url, `{"email": "login@teste.com", "password": "Senha@123"}`)
		if status != http.StatusOK {
			t.Fatalf("status = %d; want %d (body %q)", status, http.StatusOK, body)
		}

		if sub := tokenSubject(t, tokens, body); sub != testutil.FakeUserID {
			t.Errorf("token sub = %q; want %q (the user created on signup)", sub, testutil.FakeUserID)
		}
	})

	t.Run("email com outra caixa e espaços encontra a conta", func(t *testing.T) {
		t.Parallel()

		status, body := postJSON(t, url, `{"email": "  Login@Teste.COM ", "password": "Senha@123"}`)
		if status != http.StatusOK {
			t.Errorf("status = %d; want %d (body %q)", status, http.StatusOK, body)
		}
	})

	t.Run("senha errada e email inexistente respondem igual", func(t *testing.T) {
		t.Parallel()

		wrongStatus, wrongBody := postJSON(t, url, `{"email": "login@teste.com", "password": "Errada@123"}`)
		unknownStatus, unknownBody := postJSON(t, url, `{"email": "ninguem@teste.com", "password": "Senha@123"}`)

		if wrongStatus != http.StatusUnauthorized {
			t.Errorf("wrong password status = %d; want %d", wrongStatus, http.StatusUnauthorized)
		}
		if unknownStatus != http.StatusUnauthorized {
			t.Errorf("unknown email status = %d; want %d", unknownStatus, http.StatusUnauthorized)
		}
		// Respostas diferentes revelariam quais emails têm conta.
		if wrongBody != unknownBody {
			t.Errorf("bodies differ: wrong password %q, unknown email %q; want the same", wrongBody, unknownBody)
		}
		if got := errorMessage(t, wrongBody); got != "invalid credentials" {
			t.Errorf("error = %q; want %q", got, "invalid credentials")
		}
	})

	t.Run("email com formato inválido", func(t *testing.T) {
		t.Parallel()

		status, _ := postJSON(t, url, `{"email": "login.teste.com", "password": "Senha@123"}`)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d; want %d", status, http.StatusBadRequest)
		}
	})

	t.Run("método diferente de POST", func(t *testing.T) {
		t.Parallel()

		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("GET %s: %v", url, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusMethodNotAllowed)
		}
	})
}
