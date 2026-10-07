//go:build integration

package repository

import (
	"auth/internal/database"
	"auth/internal/testutil"
	"context"
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {
	ctr, cfg := testutil.StartPostgres(t)

	pool, err := database.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	t.Cleanup(pool.Close)

	h := NewHealth(pool)

	t.Run("banco disponível", func(t *testing.T) {
		if err := h.Check(context.Background()); err != nil {
			t.Errorf("Check() error = %v; want nil", err)
		}
	})

	t.Run("banco parado", func(t *testing.T) {
		if err := ctr.Stop(context.Background(), nil); err != nil {
			t.Fatalf("stop container: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.Check(ctx); err == nil {
			t.Error("Check() error = nil; want error")
		}
	})
}
