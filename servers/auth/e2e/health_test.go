//go:build integration

package e2e

import (
	"auth/internal/database"
	"auth/internal/router"
	"auth/internal/testutil"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	ctr, cfg := testutil.StartPostgres(t)

	pool, err := database.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	t.Cleanup(pool.Close)

	srv := httptest.NewServer(router.New(router.Deps{DB: pool}))
	t.Cleanup(srv.Close)

	t.Run("banco disponível", func(t *testing.T) {
		status, body := getJSON(t, srv.URL+"/api/health")

		if status != http.StatusOK {
			t.Errorf("status = %d; want %d", status, http.StatusOK)
		}
		want := map[string]any{
			"status":  "UP",
			"version": "dev",
			"checks":  map[string]any{"database": "UP"},
		}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %v; want %v", body, want)
		}
	})

	t.Run("método não permitido", func(t *testing.T) {
		resp, err := http.Post(srv.URL+"/api/health", "application/json", nil)
		if err != nil {
			t.Fatalf("POST: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusMethodNotAllowed)
		}
	})

	t.Run("rota inexistente", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/api/nada")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	// Por último: para o container, os casos acima dependem dele.
	t.Run("banco parado", func(t *testing.T) {
		if err := ctr.Stop(context.Background(), nil); err != nil {
			t.Fatalf("stop container: %v", err)
		}

		status, body := getJSON(t, srv.URL+"/api/health")

		if status != http.StatusOK {
			t.Errorf("status = %d; want %d", status, http.StatusOK)
		}
		want := map[string]any{
			"status":  "UP",
			"version": "dev",
			"checks":  map[string]any{"database": "DOWN"},
		}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %v; want %v", body, want)
		}
	})
}

func getJSON(t *testing.T, url string) (int, map[string]any) {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	return resp.StatusCode, body
}
