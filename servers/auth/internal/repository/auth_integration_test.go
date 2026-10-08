//go:build integration

package repository

import (
	"auth/internal/database"
	"auth/internal/testutil"
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	_, cfg := testutil.StartPostgres(t)
	testutil.MigrateUp(t, cfg)

	pool, err := database.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}

func TestCreateAuth(t *testing.T) {
	t.Run("Create Auth", func(t *testing.T) {
		pool := newTestPool(t)

		ctx := context.Background()

		a := NewAuthRepository(pool)

		expect := AuthData{
			ID:           "01a11c10-8e01-736d-9c09-a76d396bbc2f",
			Email:        "teste@teste.com",
			PasswordHash: "$2a$10$fakehash",
			UserID:       "01a11ce9-8e01-736d-9c09-a76d396bbc2f",
		}

		err := a.CreateAuth(ctx, expect)

		if err != nil {
			t.Fatalf("CreateAuth() unexpected error: %v", err)
		}

		var (
			gotID           string
			gotUserID       string
			gotEmail        string
			gotPasswordHash string
			gotCreatedAt    time.Time
		)

		err = pool.QueryRow(
			ctx,
			"SELECT a.id, a.user_id, a.email, a.password_hash, a.created_at FROM auth a WHERE a.id = $1",
			expect.ID,
		).Scan(&gotID, &gotUserID, &gotEmail, &gotPasswordHash, &gotCreatedAt)

		if err != nil {
			t.Fatalf("Query unexpect error: %v", err)
		}

		if gotID != expect.ID {
			t.Errorf("id = %q, want %q", gotID, expect.ID)
		}
		if gotUserID != expect.UserID {
			t.Errorf("user_id = %q, want %q", gotUserID, expect.UserID)
		}
		if gotEmail != expect.Email {
			t.Errorf("email = %q, want %q", gotEmail, expect.Email)
		}
		if gotPasswordHash != expect.PasswordHash {
			t.Errorf("password_hash = %q, want %q", gotPasswordHash, expect.PasswordHash)
		}
		if gotCreatedAt.IsZero() {
			t.Error("created_at is zero, want it set by the database")
		}

	})
}
