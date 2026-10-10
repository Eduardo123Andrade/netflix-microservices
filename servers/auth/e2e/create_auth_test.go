//go:build integration

package e2e

import (
	"auth/internal/database"
	"auth/internal/router"
	"auth/internal/testutil"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
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

	// Servidor users falso numa porta TCP real; o auth conecta nele com o
	// mesmo conector da produção.
	usersAddr := testutil.StartFakeUsers(t, &testutil.FakeUsersServer{})
	users := testutil.DialUsers(t, usersAddr)

	// Custo mínimo do bcrypt: o teste fica rápido.
	srv := httptest.NewServer(router.New(router.Deps{DB: pool, Cost: 4, Users: users}))
	t.Cleanup(srv.Close)

	url := srv.URL + createAuthPath

	t.Run("cadastro gravado no banco", func(t *testing.T) {
		t.Parallel()

		status, body := postJSON(t, url, `{
			"email": "e2e1@teste.com",
			"name": "Eduardo",
			"password": "Senha@123"
		}`)

		if status != http.StatusCreated {
			t.Fatalf("status = %d; want %d (body %q)", status, http.StatusCreated, body)
		}
		if body != "" {
			t.Errorf("body = %q; want empty", body)
		}

		if n := countAuthByEmail(t, pool, "e2e1@teste.com"); n != 1 {
			t.Fatalf("rows with email = %d; want 1", n)
		}

		var hash, userID string
		err := pool.QueryRow(context.Background(),
			"SELECT password_hash, user_id FROM auth WHERE email = $1", "e2e1@teste.com",
		).Scan(&hash, &userID)
		if err != nil {
			t.Fatalf("select auth: %v", err)
		}
		if userID != testutil.FakeUserID {
			t.Errorf("user_id = %q; want %q (the ID returned by the users service)", userID, testutil.FakeUserID)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("Senha@123")); err != nil {
			t.Errorf("password_hash = %q; want a bcrypt hash of the password: %v", hash, err)
		}
		if cost, _ := bcrypt.Cost([]byte(hash)); cost != 4 {
			t.Errorf("bcrypt cost = %d; want 4 (the configured cost)", cost)
		}
	})

	t.Run("email repetido", func(t *testing.T) {
		t.Parallel()

		body := `{
			"email": "e2e2@teste.com",
			"name": "Eduardo",
			"password": "Senha@123"
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
			"name": "Eduardo",
			"password": "Senha@123"
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

	// Usa outro servidor HTTP, apontando para um endereço sem nada escutando
	// (como o users caído de verdade), sobre o mesmo banco.
	t.Run("users fora do ar não grava credencial", func(t *testing.T) {
		t.Parallel()

		down := testutil.DialUsers(t, closedAddr(t))
		srvDown := httptest.NewServer(router.New(router.Deps{DB: pool, Cost: 4, Users: down}))
		t.Cleanup(srvDown.Close)

		code, resp := postJSON(t, srvDown.URL+createAuthPath, `{
			"email": "e2e4@teste.com",
			"name": "Eduardo",
			"password": "Senha@123"
		}`)

		if code != http.StatusInternalServerError {
			t.Errorf("status = %d; want %d", code, http.StatusInternalServerError)
		}
		if got := errorMessage(t, resp); got != "internal error" {
			t.Errorf("error = %q; want %q", got, "internal error")
		}
		if n := countAuthByEmail(t, pool, "e2e4@teste.com"); n != 0 {
			t.Errorf("rows with email = %d; want 0", n)
		}
	})
}

// closedAddr devolve um endereço local em que nada escuta: abre uma porta
// livre e fecha em seguida.
func closedAddr(t *testing.T) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := lis.Addr().String()
	lis.Close()

	return addr
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
