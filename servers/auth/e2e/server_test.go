//go:build integration

package e2e

import (
	"auth/internal/router"
	"net/http/httptest"
	"testing"
)

// newServer sobe o router real num servidor HTTP de teste e o fecha no fim.
func newServer(t *testing.T, deps router.Deps) *httptest.Server {
	t.Helper()

	h, err := router.New(deps)
	if err != nil {
		t.Fatalf("router.New() unexpected error: %v", err)
	}

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	return srv
}
