//go:build integration

package repository

import (
	"auth/internal/entity"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	rtUserID   = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
	rtFamilyA  = "01a11d00-8e01-736d-9c09-a76d396bbc2f"
	rtFamilyB  = "01a11d01-8e01-736d-9c09-a76d396bbc2f"
	rtTokenA1  = "01a11e00-8e01-736d-9c09-a76d396bbc2f"
	rtTokenA2  = "01a11e01-8e01-736d-9c09-a76d396bbc2f"
	rtTokenB1  = "01a11e02-8e01-736d-9c09-a76d396bbc2f"
	rtTokenOne = "01a11e03-8e01-736d-9c09-a76d396bbc2f"
)

// rtBase é truncado no segundo: o Postgres guarda microssegundos, e assim a
// comparação com o que volta do banco não depende de arredondamento.
var rtBase = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

func newRefreshToken(id, familyID, hash string) entity.RefreshToken {
	return entity.RefreshToken{
		ID:        id,
		FamilyID:  familyID,
		UserID:    rtUserID,
		TokenHash: hash,
		ExpiresAt: rtBase.Add(30 * 24 * time.Hour),
	}
}

func createRefreshToken(t *testing.T, r *RefreshTokenRepository, rt entity.RefreshToken) {
	t.Helper()

	if err := r.Create(context.Background(), rt); err != nil {
		t.Fatalf("Create(%s) unexpected error: %v", rt.ID, err)
	}
}

func findRefreshToken(t *testing.T, r *RefreshTokenRepository, hash string) entity.RefreshToken {
	t.Helper()

	got, err := r.FindByHash(context.Background(), hash)
	if err != nil {
		t.Fatalf("FindByHash(%s) unexpected error: %v", hash, err)
	}
	return got
}

// nullColumns diz se rotated_at e revoked_at estão NULL de verdade no banco
// (o zero de time.Time não pode virar "0001-01-01").
func nullColumns(t *testing.T, pool *pgxpool.Pool, id string) (rotatedNull, revokedNull bool) {
	t.Helper()

	err := pool.QueryRow(context.Background(),
		"SELECT rotated_at IS NULL, revoked_at IS NULL FROM refresh_tokens WHERE id = $1", id,
	).Scan(&rotatedNull, &revokedNull)
	if err != nil {
		t.Fatalf("select refresh token %s: %v", id, err)
	}
	return rotatedNull, revokedNull
}

func TestRefreshTokenRepository(t *testing.T) {
	t.Parallel()
	pool := newTestPool(t)

	t.Run("Create and find by hash", func(t *testing.T) {
		r := NewRefreshTokenRepository(pool)
		want := newRefreshToken(rtTokenOne, rtFamilyB, "hash-create-find")

		createRefreshToken(t, r, want)
		got := findRefreshToken(t, r, want.TokenHash)

		if got.ID != want.ID || got.FamilyID != want.FamilyID || got.UserID != want.UserID || got.TokenHash != want.TokenHash {
			t.Errorf("FindByHash() = %+v, want %+v", got, want)
		}
		if !got.ExpiresAt.Equal(want.ExpiresAt) {
			t.Errorf("expires_at = %v, want %v", got.ExpiresAt, want.ExpiresAt)
		}
		if !got.RotatedAt.IsZero() {
			t.Errorf("rotated_at = %v, want zero (never rotated)", got.RotatedAt)
		}
		if !got.RevokedAt.IsZero() {
			t.Errorf("revoked_at = %v, want zero (never revoked)", got.RevokedAt)
		}

		if rotatedNull, revokedNull := nullColumns(t, pool, want.ID); !rotatedNull || !revokedNull {
			t.Errorf("rotated_at NULL = %v, revoked_at NULL = %v; want both NULL in the database", rotatedNull, revokedNull)
		}
	})

	t.Run("Error duplicated hash", func(t *testing.T) {
		r := NewRefreshTokenRepository(pool)
		first := newRefreshToken("01a11e10-8e01-736d-9c09-a76d396bbc2f", rtFamilyB, "hash-duplicated")
		second := newRefreshToken("01a11e11-8e01-736d-9c09-a76d396bbc2f", rtFamilyB, "hash-duplicated")

		createRefreshToken(t, r, first)
		if err := r.Create(context.Background(), second); err == nil {
			t.Error("Create() = error nil; want error for a duplicated token hash")
		}
	})

	t.Run("Error hash not found", func(t *testing.T) {
		r := NewRefreshTokenRepository(pool)

		_, err := r.FindByHash(context.Background(), "hash-that-does-not-exist")
		if !errors.Is(err, ErrRefreshTokenNotFound) {
			t.Errorf("FindByHash() = error %v; want %v", err, ErrRefreshTokenNotFound)
		}
	})

	t.Run("Mark rotated keeps the first rotation time", func(t *testing.T) {
		r := NewRefreshTokenRepository(pool)
		rt := newRefreshToken("01a11e20-8e01-736d-9c09-a76d396bbc2f", rtFamilyB, "hash-rotate")
		other := newRefreshToken("01a11e21-8e01-736d-9c09-a76d396bbc2f", rtFamilyB, "hash-rotate-other")
		createRefreshToken(t, r, rt)
		createRefreshToken(t, r, other)

		first := rtBase.Add(time.Minute)
		if err := r.MarkRotated(context.Background(), rt.ID, first); err != nil {
			t.Fatalf("MarkRotated() unexpected error: %v", err)
		}
		// A janela de tolerância conta a partir da primeira rotação: um segundo
		// uso dentro dela não pode empurrar o prazo para frente.
		if err := r.MarkRotated(context.Background(), rt.ID, first.Add(3*time.Minute)); err != nil {
			t.Fatalf("second MarkRotated() unexpected error: %v", err)
		}

		if got := findRefreshToken(t, r, rt.TokenHash); !got.RotatedAt.Equal(first) {
			t.Errorf("rotated_at = %v, want %v (the first rotation)", got.RotatedAt, first)
		}
		if got := findRefreshToken(t, r, other.TokenHash); !got.RotatedAt.IsZero() {
			t.Errorf("other token rotated_at = %v, want zero (only the given ID is rotated)", got.RotatedAt)
		}
	})

	t.Run("Revoke family revokes only that family", func(t *testing.T) {
		r := NewRefreshTokenRepository(pool)
		a1 := newRefreshToken(rtTokenA1, rtFamilyA, "hash-family-a1")
		a2 := newRefreshToken(rtTokenA2, rtFamilyA, "hash-family-a2")
		b1 := newRefreshToken(rtTokenB1, rtFamilyB, "hash-family-b1")
		for _, rt := range []entity.RefreshToken{a1, a2, b1} {
			createRefreshToken(t, r, rt)
		}

		at := rtBase.Add(10 * time.Minute)
		if err := r.RevokeFamily(context.Background(), rtFamilyA, at); err != nil {
			t.Fatalf("RevokeFamily() unexpected error: %v", err)
		}

		for _, rt := range []entity.RefreshToken{a1, a2} {
			if got := findRefreshToken(t, r, rt.TokenHash); !got.RevokedAt.Equal(at) {
				t.Errorf("token %s revoked_at = %v, want %v", rt.ID, got.RevokedAt, at)
			}
		}
		if got := findRefreshToken(t, r, b1.TokenHash); !got.RevokedAt.IsZero() {
			t.Errorf("token of another family revoked_at = %v, want zero", got.RevokedAt)
		}
	})

	t.Run("Revoke family keeps the first revocation time", func(t *testing.T) {
		r := NewRefreshTokenRepository(pool)
		family := "01a11d02-8e01-736d-9c09-a76d396bbc2f"
		rt := newRefreshToken("01a11e30-8e01-736d-9c09-a76d396bbc2f", family, "hash-revoke-twice")
		createRefreshToken(t, r, rt)

		first := rtBase.Add(time.Minute)
		if err := r.RevokeFamily(context.Background(), family, first); err != nil {
			t.Fatalf("RevokeFamily() unexpected error: %v", err)
		}
		if err := r.RevokeFamily(context.Background(), family, first.Add(time.Hour)); err != nil {
			t.Fatalf("second RevokeFamily() unexpected error: %v", err)
		}

		if got := findRefreshToken(t, r, rt.TokenHash); !got.RevokedAt.Equal(first) {
			t.Errorf("revoked_at = %v, want %v (the first revocation)", got.RevokedAt, first)
		}
	})

	t.Run("Revoke unknown family is not an error", func(t *testing.T) {
		r := NewRefreshTokenRepository(pool)

		// Logout é idempotente: revogar algo que já não existe não falha.
		if err := r.RevokeFamily(context.Background(), "01a11dff-8e01-736d-9c09-a76d396bbc2f", rtBase); err != nil {
			t.Errorf("RevokeFamily() unexpected error: %v", err)
		}
	})

	t.Run("Error context canceled", func(t *testing.T) {
		r := NewRefreshTokenRepository(pool)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if err := r.Create(ctx, newRefreshToken("01a11e40-8e01-736d-9c09-a76d396bbc2f", rtFamilyB, "hash-canceled")); !errors.Is(err, context.Canceled) {
			t.Errorf("Create() = error %v; want %v", err, context.Canceled)
		}
		if _, err := r.FindByHash(ctx, "hash-create-find"); !errors.Is(err, context.Canceled) {
			t.Errorf("FindByHash() = error %v; want %v", err, context.Canceled)
		}
	})
}
