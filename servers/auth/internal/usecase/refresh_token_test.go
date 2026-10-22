package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/utils"
)

const refreshTTL = 30 * 24 * time.Hour

// refreshNow é o "agora" fixo dos testes; cada caso monta o relógio a partir dele.
var refreshNow = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

func clockAt(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// hashOf é como o usecase precisa guardar o token: SHA-256 em hexadecimal.
func hashOf(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// fakeRefreshStore substitui o repositório de refresh tokens: guarda os
// tokens num mapa por hash, devolve os erros configurados e registra as chamadas.
type fakeRefreshStore struct {
	tokens map[string]entity.RefreshToken

	createErr error
	findErr   error
	rotateErr error
	revokeErr error

	created       []entity.RefreshToken
	findCalled    bool
	rotatedID     string
	rotatedAt     time.Time
	rotateCalls   int
	revokedFamily string
	revokedAt     time.Time
	revokeCalls   int
}

func newFakeRefreshStore(tokens ...entity.RefreshToken) *fakeRefreshStore {
	f := &fakeRefreshStore{tokens: map[string]entity.RefreshToken{}}
	for _, rt := range tokens {
		f.tokens[rt.TokenHash] = rt
	}
	return f
}

func (f *fakeRefreshStore) Create(ctx context.Context, rt entity.RefreshToken) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, rt)
	f.tokens[rt.TokenHash] = rt
	return nil
}

func (f *fakeRefreshStore) FindByHash(ctx context.Context, hash string) (entity.RefreshToken, error) {
	f.findCalled = true
	if f.findErr != nil {
		return entity.RefreshToken{}, f.findErr
	}
	rt, ok := f.tokens[hash]
	if !ok {
		return entity.RefreshToken{}, repository.ErrRefreshTokenNotFound
	}
	return rt, nil
}

func (f *fakeRefreshStore) MarkRotated(ctx context.Context, id string, at time.Time) error {
	f.rotateCalls++
	f.rotatedID = id
	f.rotatedAt = at
	return f.rotateErr
}

func (f *fakeRefreshStore) RevokeFamily(ctx context.Context, familyID string, at time.Time) error {
	f.revokeCalls++
	f.revokedFamily = familyID
	f.revokedAt = at
	return f.revokeErr
}

const (
	existingPlain  = "existing-refresh-token"
	existingID     = "01a11e00-8e01-736d-9c09-a76d396bbc2f"
	existingFamily = "01a11d00-8e01-736d-9c09-a76d396bbc2f"
)

// existingToken é um refresh token válido, emitido há um dia.
func existingToken() entity.RefreshToken {
	return entity.RefreshToken{
		ID:        existingID,
		FamilyID:  existingFamily,
		UserID:    fakeUserID,
		TokenHash: hashOf(existingPlain),
		ExpiresAt: refreshNow.Add(refreshTTL - 24*time.Hour),
	}
}

// wantOpaqueToken confere o formato do token entregue ao cliente: aleatório,
// com pelo menos 32 bytes, em base64 URL-safe sem padding (vai em JSON e em URL).
func wantOpaqueToken(t *testing.T, plain string) {
	t.Helper()

	raw, err := base64.RawURLEncoding.DecodeString(plain)
	if err != nil {
		t.Fatalf("token %q is not base64 URL-safe without padding: %v", plain, err)
	}
	if len(raw) < 32 {
		t.Errorf("token has %d random bytes, want at least 32", len(raw))
	}
}

// wantIssued confere o token gravado a partir do valor entregue ao cliente.
func wantIssued(t *testing.T, got entity.RefreshToken, plain, userID string) {
	t.Helper()

	if got.TokenHash != hashOf(plain) {
		t.Errorf("token_hash = %q, want the SHA-256 hex of the returned token", got.TokenHash)
	}
	if got.TokenHash == plain {
		t.Error("token_hash equals the plain token, want only the hash stored")
	}
	if got.UserID != userID {
		t.Errorf("user_id = %q, want %q", got.UserID, userID)
	}
	if !utils.IsUUIDV7(got.ID) {
		t.Errorf("id = %q, want a UUID v7", got.ID)
	}
	if want := refreshNow.Add(refreshTTL); !got.ExpiresAt.Equal(want) {
		t.Errorf("expires_at = %v, want %v (now + ttl)", got.ExpiresAt, want)
	}
	if !got.RotatedAt.IsZero() || !got.RevokedAt.IsZero() {
		t.Errorf("rotated_at = %v, revoked_at = %v; want both zero on a new token", got.RotatedAt, got.RevokedAt)
	}
}

func TestRefreshTokenIssue(t *testing.T) {
	t.Parallel()

	t.Run("Issues a token in a new family", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore()
		uc := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow))

		plain, err := uc.Issue(context.Background(), fakeUserID)
		if err != nil {
			t.Fatalf("Issue() unexpected error: %v", err)
		}

		wantOpaqueToken(t, plain)
		if len(store.created) != 1 {
			t.Fatalf("Create() called %d times, want 1", len(store.created))
		}
		got := store.created[0]
		wantIssued(t, got, plain, fakeUserID)
		if !utils.IsUUIDV7(got.FamilyID) {
			t.Errorf("family_id = %q, want a UUID v7", got.FamilyID)
		}
	})

	t.Run("Each login gets its own token and family", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore()
		uc := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow))

		first, err := uc.Issue(context.Background(), fakeUserID)
		if err != nil {
			t.Fatalf("first Issue() unexpected error: %v", err)
		}
		second, err := uc.Issue(context.Background(), fakeUserID)
		if err != nil {
			t.Fatalf("second Issue() unexpected error: %v", err)
		}

		if first == second {
			t.Error("both Issue() calls returned the same token, want random tokens")
		}
		// Família por login: o logout de um dispositivo não derruba os outros.
		if store.created[0].FamilyID == store.created[1].FamilyID {
			t.Error("both tokens share a family, want a new family per login")
		}
	})

	t.Run("Error saving", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore()
		dbErr := errors.New("connection refused")
		store.createErr = dbErr

		plain, err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Issue(context.Background(), fakeUserID)
		if !errors.Is(err, dbErr) {
			t.Errorf("Issue() = error %v; want it to wrap %v", err, dbErr)
		}
		if plain != "" {
			t.Errorf("Issue() token = %q, want empty on error (it was not saved)", plain)
		}
	})
}

func TestRefreshTokenRotate(t *testing.T) {
	t.Parallel()

	t.Run("Rotates a valid token", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore(existingToken())
		uc := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow))

		userID, plain, err := uc.Rotate(context.Background(), existingPlain)
		if err != nil {
			t.Fatalf("Rotate() unexpected error: %v", err)
		}

		if userID != fakeUserID {
			t.Errorf("Rotate() userID = %q, want %q", userID, fakeUserID)
		}
		wantOpaqueToken(t, plain)
		if plain == existingPlain {
			t.Error("Rotate() returned the same token, want a new one")
		}

		if store.rotatedID != existingID || !store.rotatedAt.Equal(refreshNow) {
			t.Errorf("MarkRotated() = (%q, %v), want (%q, %v)", store.rotatedID, store.rotatedAt, existingID, refreshNow)
		}
		if len(store.created) != 1 {
			t.Fatalf("Create() called %d times, want 1", len(store.created))
		}
		got := store.created[0]
		wantIssued(t, got, plain, fakeUserID)
		if got.FamilyID != existingFamily {
			t.Errorf("family_id = %q, want %q (rotation stays in the same family)", got.FamilyID, existingFamily)
		}
		if store.revokeCalls != 0 {
			t.Error("RevokeFamily() called, want no revocation on a normal rotation")
		}
	})

	t.Run("Rejects invalid tokens", func(t *testing.T) {
		t.Parallel()

		expired := existingToken()
		expired.ExpiresAt = refreshNow

		revoked := existingToken()
		revoked.RevokedAt = refreshNow.Add(-time.Hour)

		tests := []struct {
			name  string
			token entity.RefreshToken
			plain string
		}{
			{name: "unknown token", token: existingToken(), plain: "token-that-was-never-issued"},
			// expires_at é o primeiro instante em que o token já não vale.
			{name: "expired token", token: expired, plain: existingPlain},
			{name: "revoked token", token: revoked, plain: existingPlain},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				store := newFakeRefreshStore(tt.token)
				userID, plain, err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Rotate(context.Background(), tt.plain)

				if err != ErrInvalidRefreshToken {
					t.Errorf("Rotate() = error %v; want exactly %v", err, ErrInvalidRefreshToken)
				}
				if userID != "" || plain != "" {
					t.Errorf("Rotate() = (%q, %q), want empty values on error", userID, plain)
				}
				if len(store.created) != 0 {
					t.Error("Create() called, want no new token for an invalid one")
				}
				if store.rotateCalls != 0 {
					t.Error("MarkRotated() called, want nothing changed for an invalid token")
				}
			})
		}
	})

	t.Run("Empty token is rejected without touching the database", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore()
		_, _, err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Rotate(context.Background(), "")

		if err != ErrInvalidRefreshToken {
			t.Errorf("Rotate() = error %v; want exactly %v", err, ErrInvalidRefreshToken)
		}
		if store.findCalled {
			t.Error("FindByHash() called, want empty token rejected before the database")
		}
	})

	t.Run("Accepts a rotated token inside the grace period", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			since time.Duration
		}{
			{name: "just after rotation", since: time.Second},
			{name: "exactly at the end of the grace period", since: refreshGracePeriod},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				// Dois pedidos quase juntos (duas abas, retry depois de timeout) usam o
				// mesmo token: o segundo não pode ser tratado como roubo.
				rotated := existingToken()
				rotated.RotatedAt = refreshNow.Add(-tt.since)
				store := newFakeRefreshStore(rotated)

				userID, plain, err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Rotate(context.Background(), existingPlain)
				if err != nil {
					t.Fatalf("Rotate() unexpected error: %v", err)
				}

				if userID != fakeUserID {
					t.Errorf("Rotate() userID = %q, want %q", userID, fakeUserID)
				}
				if len(store.created) != 1 {
					t.Fatalf("Create() called %d times, want 1", len(store.created))
				}
				wantIssued(t, store.created[0], plain, fakeUserID)
				if store.created[0].FamilyID != existingFamily {
					t.Errorf("family_id = %q, want %q", store.created[0].FamilyID, existingFamily)
				}
				if store.revokeCalls != 0 {
					t.Error("RevokeFamily() called, want no revocation inside the grace period")
				}
			})
		}
	})

	t.Run("Reuse after the grace period revokes the family", func(t *testing.T) {
		t.Parallel()

		// Um token girado reaparecendo depois da janela indica que ele vazou:
		// quem quer que esteja com a família perde o acesso.
		rotated := existingToken()
		rotated.RotatedAt = refreshNow.Add(-refreshGracePeriod - time.Second)
		store := newFakeRefreshStore(rotated)

		userID, plain, err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Rotate(context.Background(), existingPlain)

		// Para o cliente é o mesmo "token inválido"...
		if !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Rotate() = error %v; want it to include %v", err, ErrInvalidRefreshToken)
		}
		// ...mas o handler precisa saber que foi reuso, para registrar no log.
		if !errors.Is(err, ErrRefreshTokenReused) {
			t.Errorf("Rotate() = error %v; want it to include %v", err, ErrRefreshTokenReused)
		}
		if userID != "" || plain != "" {
			t.Errorf("Rotate() = (%q, %q), want empty values on reuse", userID, plain)
		}
		if store.revokedFamily != existingFamily || !store.revokedAt.Equal(refreshNow) {
			t.Errorf("RevokeFamily() = (%q, %v), want (%q, %v)", store.revokedFamily, store.revokedAt, existingFamily, refreshNow)
		}
		if len(store.created) != 0 {
			t.Error("Create() called, want no new token on reuse")
		}
	})

	t.Run("Error finding token is not treated as invalid token", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore()
		dbErr := errors.New("connection refused")
		store.findErr = dbErr

		_, _, err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Rotate(context.Background(), existingPlain)
		if !errors.Is(err, dbErr) {
			t.Errorf("Rotate() = error %v; want it to wrap %v", err, dbErr)
		}
		// Banco fora não é culpa do cliente: vira 500, não 401.
		if errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Rotate() = %v; want an error other than invalid refresh token", err)
		}
	})

	t.Run("Error marking rotated does not issue a new token", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore(existingToken())
		dbErr := errors.New("connection refused")
		store.rotateErr = dbErr

		userID, plain, err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Rotate(context.Background(), existingPlain)
		if !errors.Is(err, dbErr) {
			t.Errorf("Rotate() = error %v; want it to wrap %v", err, dbErr)
		}
		if userID != "" || plain != "" {
			t.Errorf("Rotate() = (%q, %q), want empty values on error", userID, plain)
		}
		// Sem marcar o antigo, um token novo deixaria dois válidos na família.
		if len(store.created) != 0 {
			t.Error("Create() called, want no new token when the old one could not be marked")
		}
	})

	t.Run("Error saving the new token", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore(existingToken())
		dbErr := errors.New("connection refused")
		store.createErr = dbErr

		userID, plain, err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Rotate(context.Background(), existingPlain)
		if !errors.Is(err, dbErr) {
			t.Errorf("Rotate() = error %v; want it to wrap %v", err, dbErr)
		}
		if errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Rotate() = %v; want an error other than invalid refresh token", err)
		}
		if userID != "" || plain != "" {
			t.Errorf("Rotate() = (%q, %q), want empty values on error", userID, plain)
		}
	})
}

func TestRefreshTokenRevoke(t *testing.T) {
	t.Parallel()

	t.Run("Revokes the token family", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore(existingToken())

		if err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Revoke(context.Background(), existingPlain); err != nil {
			t.Fatalf("Revoke() unexpected error: %v", err)
		}

		// Revoga a família, não só o token: os girados dentro da janela também param.
		if store.revokedFamily != existingFamily || !store.revokedAt.Equal(refreshNow) {
			t.Errorf("RevokeFamily() = (%q, %v), want (%q, %v)", store.revokedFamily, store.revokedAt, existingFamily, refreshNow)
		}
	})

	t.Run("Revokes the family even from a rotated token", func(t *testing.T) {
		t.Parallel()

		rotated := existingToken()
		rotated.RotatedAt = refreshNow.Add(-time.Hour)
		store := newFakeRefreshStore(rotated)

		if err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Revoke(context.Background(), existingPlain); err != nil {
			t.Fatalf("Revoke() unexpected error: %v", err)
		}
		if store.revokedFamily != existingFamily {
			t.Errorf("RevokeFamily() family = %q, want %q", store.revokedFamily, existingFamily)
		}
	})

	t.Run("Unknown token is not an error", func(t *testing.T) {
		t.Parallel()

		// Logout é idempotente: o cliente apaga o token de qualquer jeito, e
		// responder diferente diria se um token existe.
		store := newFakeRefreshStore()

		if err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Revoke(context.Background(), "token-that-was-never-issued"); err != nil {
			t.Errorf("Revoke() unexpected error: %v", err)
		}
		if store.revokeCalls != 0 {
			t.Error("RevokeFamily() called, want nothing to revoke for an unknown token")
		}
	})

	t.Run("Empty token is not an error and skips the database", func(t *testing.T) {
		t.Parallel()

		store := newFakeRefreshStore()

		if err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Revoke(context.Background(), ""); err != nil {
			t.Errorf("Revoke() unexpected error: %v", err)
		}
		if store.findCalled {
			t.Error("FindByHash() called, want empty token skipped before the database")
		}
	})

	t.Run("Errors from the database are returned", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("connection refused")

		tests := []struct {
			name  string
			setup func(*fakeRefreshStore)
		}{
			{name: "finding", setup: func(f *fakeRefreshStore) { f.findErr = dbErr }},
			{name: "revoking", setup: func(f *fakeRefreshStore) { f.revokeErr = dbErr }},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				store := newFakeRefreshStore(existingToken())
				tt.setup(store)

				err := NewRefreshTokenUseCase(store, refreshTTL, clockAt(refreshNow)).Revoke(context.Background(), existingPlain)
				if !errors.Is(err, dbErr) {
					t.Errorf("Revoke() = error %v; want it to wrap %v", err, dbErr)
				}
			})
		}
	})
}
