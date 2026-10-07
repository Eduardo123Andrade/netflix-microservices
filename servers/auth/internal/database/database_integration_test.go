//go:build integration

package database

import (
	"auth/internal/config"
	"auth/internal/testutil"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConnect(t *testing.T) {
	_, cfg := testutil.StartPostgres(t)

	t.Run("banco disponível", func(t *testing.T) {
		pool := newTestPool(t, cfg)

		if err := Connect(context.Background(), pool); err != nil {
			t.Errorf("Connect() error = %v; want nil", err)
		}
	})

	t.Run("porta sem banco", func(t *testing.T) {
		pool := newTestPool(t, config.Database{
			User:     "x",
			Password: "x",
			Name:     "x",
			Host:     "127.0.0.1",
			Port:     1,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := Connect(ctx, pool); !errors.Is(err, ErrPing) {
			t.Errorf("Connect() error = %v; want %v", err, ErrPing)
		}
	})

	t.Run("contexto cancelado", func(t *testing.T) {
		pool := newTestPool(t, cfg)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if err := Connect(ctx, pool); !errors.Is(err, ErrPing) {
			t.Errorf("Connect() error = %v; want %v", err, ErrPing)
		}
	})
}

func newTestPool(t *testing.T, cfg config.Database) *pgxpool.Pool {
	t.Helper()

	pool, err := NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}
