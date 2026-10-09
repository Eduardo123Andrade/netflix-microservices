//go:build integration

package repository

import (
	"auth/internal/database"
	"auth/internal/testutil"
	"context"
	"errors"
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

func insertAuth(t *testing.T, pool *pgxpool.Pool, data AuthData) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		"INSERT INTO auth (id, email, password_hash, user_id) VALUES ($1, $2, $3, $4)",
		data.ID, data.Email, data.PasswordHash, data.UserID,
	)
	if err != nil {
		t.Fatalf("insertAuth(%s) unexpected error: %v", data.Email, err)
	}
}

func TestCreateAuth(t *testing.T) {
	t.Parallel()
	pool := newTestPool(t)

	t.Run("Create Auth", func(t *testing.T) {
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

	t.Run("Error email already exists", func(t *testing.T) {
		a := NewAuthRepository(pool)
		ctx := context.Background()

		expect := AuthData{
			ID:           "01a11c11-8e01-736d-9c09-a76d396bbc2f",
			Email:        "teste2@teste.com",
			PasswordHash: "$2a$10$fakehash",
			UserID:       "01a11ce9-8e01-736d-9c09-a76d396bbc2f",
		}

		err := a.CreateAuth(ctx, expect)

		if err != nil {
			t.Fatalf("CreateAuth() unexpected error: %v", err)
		}

		dub := expect
		dub.ID = "01a11c12-8e01-736d-9c09-a76d396bbc2f"

		err = a.CreateAuth(ctx, dub)

		if err == nil {
			t.Fatalf("CreateAuth() = error nil; want = err")
		}

		if !errors.Is(err, ErrAuthAlreadyExists) {
			t.Fatalf("CreateAuth() = error %v; want = %v", err, ErrAuthAlreadyExists)
		}
	})

	t.Run("Error context canceled", func(t *testing.T) {
		a := NewAuthRepository(pool)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := a.CreateAuth(ctx, AuthData{
			ID:           "01a11c12-8e01-736d-9c09-a76d396bbc2f",
			Email:        "teste3@teste.com",
			PasswordHash: "$2a$10$fakehash",
			UserID:       "01a11ce9-8e01-736d-9c09-a76d396bbc2f",
		})

		if err == nil {
			t.Fatal("CreateAuth() = error nil; want error")
		}
		if errors.Is(err, ErrAuthAlreadyExists) {
			t.Fatalf("CreateAuth() = %v; want a non-duplicate error", err)
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("CreateAuth() = %v; want it to wrap context.Canceled", err)
		}
	})
}
func TestFindByEmail(t *testing.T) {
	t.Parallel()
	pool := newTestPool(t)

	t.Run("Find auth", func(t *testing.T) {
		ctx := context.Background()
		a := NewAuthRepository(pool)

		want := AuthData{
			ID:           "01a11d10-8e01-736d-9c09-a76d396bbc2f",
			Email:        "find1@teste.com",
			PasswordHash: "$2a$10$fakehash",
			UserID:       "01a11de9-8e01-736d-9c09-a76d396bbc2f",
		}
		insertAuth(t, pool, want)

		got, err := a.FindByEmail(ctx, want.Email)
		if err != nil {
			t.Fatalf("FindByEmail() unexpected error: %v", err)
		}

		if got.ID != want.ID {
			t.Errorf("id = %q, want %q", got.ID, want.ID)
		}
		if got.UserID != want.UserID {
			t.Errorf("user_id = %q, want %q", got.UserID, want.UserID)
		}
		if got.Email != want.Email {
			t.Errorf("email = %q, want %q", got.Email, want.Email)
		}
		if got.PasswordHash != want.PasswordHash {
			t.Errorf("password_hash = %q, want %q", got.PasswordHash, want.PasswordHash)
		}
	})

	t.Run("Find the right auth among many", func(t *testing.T) {
		ctx := context.Background()
		a := NewAuthRepository(pool)

		first := AuthData{
			ID:           "01a11d11-8e01-736d-9c09-a76d396bbc2f",
			Email:        "find2a@teste.com",
			PasswordHash: "$2a$10$fakehash",
			UserID:       "01a11de9-8e01-736d-9c09-a76d396bbc2f",
		}
		second := AuthData{
			ID:           "01a11d12-8e01-736d-9c09-a76d396bbc2f",
			Email:        "find2b@teste.com",
			PasswordHash: "$2a$10$fakehash",
			UserID:       "01a11de9-8e01-736d-9c09-a76d396bbc2f",
		}
		insertAuth(t, pool, first)
		insertAuth(t, pool, second)

		got, err := a.FindByEmail(ctx, second.Email)
		if err != nil {
			t.Fatalf("FindByEmail() unexpected error: %v", err)
		}
		if got.ID != second.ID {
			t.Errorf("id = %q, want %q", got.ID, second.ID)
		}
	})

	t.Run("Error not found", func(t *testing.T) {
		a := NewAuthRepository(pool)

		_, err := a.FindByEmail(context.Background(), "naoexiste@teste.com")

		if !errors.Is(err, ErrAuthNotFound) {
			t.Fatalf("FindByEmail() = error %v; want %v", err, ErrAuthNotFound)
		}
	})

	t.Run("Error context canceled", func(t *testing.T) {
		a := NewAuthRepository(pool)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := a.FindByEmail(ctx, "find4@teste.com")
		if err == nil {
			t.Fatal("FindByEmail() = error nil; want error")
		}
		if errors.Is(err, ErrAuthNotFound) {
			t.Fatalf("FindByEmail() = %v; want an error other than not found", err)
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("FindByEmail() = %v; want it to wrap context.Canceled", err)
		}
	})

	t.Run("Search is exact", func(t *testing.T) {
		ctx := context.Background()
		a := NewAuthRepository(pool)

		insertAuth(t, pool, AuthData{
			ID:           "01a11d15-8e01-736d-9c09-a76d396bbc2f",
			Email:        "find5@teste.com",
			PasswordHash: "$2a$10$fakehash",
			UserID:       "01a11de9-8e01-736d-9c09-a76d396bbc2f",
		})

		_, err := a.FindByEmail(ctx, "FIND5@teste.com")
		if !errors.Is(err, ErrAuthNotFound) {
			t.Fatalf("FindByEmail() = error %v; want %v (repository does not normalize)", err, ErrAuthNotFound)
		}
	})
}
