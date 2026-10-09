//go:build integration

package e2e

import (
	"auth/internal/database"
	"auth/internal/router"
	"auth/internal/testutil"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createAuthPath = "/api/auth/create_user"

func TestCreateAuthEndpoint(t *testing.T) {
	t.Parallel()
	_, cfg := testutil.StartPostgres(t)
	testutil.MigrateUp(t, cfg)

	pool, err := database.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	t.Cleanup(pool.Close)

	srv := httptest.NewServer(router.New(router.Deps{DB: pool}))
	t.Cleanup(srv.Close)

	url := srv.URL + createAuthPath

	t.Run("cadastro gravado no banco", func(t *testing.T) {
		t.Parallel()

		status, body := postJSON(t, url, `{
			"email": "e2e1@teste.com",
			"password_hash": "$2a$10$fakehash",
			"user_id": "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
		}`)

		if status != http.StatusCreated {
			t.Fatalf("status = %d; want %d (body %q)", status, http.StatusCreated, body)
		}
		if body != "" {
			t.Errorf("body = %q; want empty", body)
		}

		if n := countAuthByEmail(t, pool, "e2e1@teste.com"); n != 1 {
			t.Errorf("rows with email = %d; want 1", n)
		}
	})

	t.Run("email repetido", func(t *testing.T) {
		t.Parallel()

		body := `{
			"email": "e2e2@teste.com",
			"password_hash": "$2a$10$fakehash",
			"user_id": "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
		}`

		if status, resp := postJSON(t, url, body); status != http.StatusCreated {
			t.Fatalf("first POST status = %d; want %d (body %q)", status, http.StatusCreated, resp)
		}

		status, resp := postJSON(t, url, body)
		if status != http.StatusConflict {
			t.Errorf("second POST status = %d; want %d", status, http.StatusConflict)
		}
		if got := errorMessage(t, resp); got != "email already registered" {
			t.Errorf("error = %q; want %q", got, "email already registered")
		}

		if n := countAuthByEmail(t, pool, "e2e2@teste.com"); n != 1 {
			t.Errorf("rows with email = %d; want 1", n)
		}
	})

	t.Run("entrada inválida não grava", func(t *testing.T) {
		t.Parallel()

		status, resp := postJSON(t, url, `{
			"email": "e2e3.teste.com",
			"password_hash": "$2a$10$fakehash",
			"user_id": "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
		}`)

		if status != http.StatusBadRequest {
			t.Errorf("status = %d; want %d", status, http.StatusBadRequest)
		}
		if got := errorMessage(t, resp); !strings.Contains(got, "invalid email") {
			t.Errorf("error = %q; want it to contain %q", got, "invalid email")
		}

		if n := countAuthByEmail(t, pool, "e2e3.teste.com"); n != 0 {
			t.Errorf("rows with email = %d; want 0", n)
		}
	})

	t.Run("corpo inválido", func(t *testing.T) {
		t.Parallel()

		status, resp := postJSON(t, url, `{"email": `)

		if status != http.StatusBadRequest {
			t.Errorf("status = %d; want %d", status, http.StatusBadRequest)
		}
		if got := errorMessage(t, resp); got != "invalid request body" {
			t.Errorf("error = %q; want %q", got, "invalid request body")
		}
	})

	t.Run("método não permitido", func(t *testing.T) {
		t.Parallel()

		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusMethodNotAllowed)
		}
	})
}

func postJSON(t *testing.T, url, body string) (int, string) {
	t.Helper()

	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return resp.StatusCode, string(raw)
}

func errorMessage(t *testing.T, body string) string {
	t.Helper()

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode error body %q: %v", body, err)
	}

	return resp.Error
}

func countAuthByEmail(t *testing.T, pool *pgxpool.Pool, email string) int {
	t.Helper()

	var n int
	err := pool.QueryRow(context.Background(),
		"SELECT count(*) FROM auth WHERE email = $1", email,
	).Scan(&n)
	if err != nil {
		t.Fatalf("count auth: %v", err)
	}

	return n
}
