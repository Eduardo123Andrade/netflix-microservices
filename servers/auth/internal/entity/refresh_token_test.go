package entity

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"auth/internal/utils"
)

func TestNewRefreshToken(t *testing.T) {
	t.Parallel()

	const (
		familyID = "01a11d00-8e01-736d-9c09-a76d396bbc2f"
		userID   = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
	)
	expiresAt := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)

	t.Run("Fills fields from arguments", func(t *testing.T) {
		t.Parallel()

		got, _, err := NewRefreshToken(familyID, userID, expiresAt)
		if err != nil {
			t.Fatalf("NewRefreshToken() unexpected error: %v", err)
		}

		if got.FamilyID != familyID {
			t.Errorf("family_id = %q, want %q", got.FamilyID, familyID)
		}
		if got.UserID != userID {
			t.Errorf("user_id = %q, want %q", got.UserID, userID)
		}
		if !got.ExpiresAt.Equal(expiresAt) {
			t.Errorf("expires_at = %v, want %v", got.ExpiresAt, expiresAt)
		}
	})

	t.Run("Starts neither rotated nor revoked", func(t *testing.T) {
		t.Parallel()

		got, _, err := NewRefreshToken(familyID, userID, expiresAt)
		if err != nil {
			t.Fatalf("NewRefreshToken() unexpected error: %v", err)
		}

		if !got.RotatedAt.IsZero() || !got.RevokedAt.IsZero() {
			t.Errorf("rotated_at = %v, revoked_at = %v; want both zero on a new token", got.RotatedAt, got.RevokedAt)
		}
	})

	t.Run("Generates a UUID v7 ID", func(t *testing.T) {
		t.Parallel()

		got, _, err := NewRefreshToken(familyID, userID, expiresAt)
		if err != nil {
			t.Fatalf("NewRefreshToken() unexpected error: %v", err)
		}

		if !utils.IsUUIDV7(got.ID) {
			t.Errorf("id = %q, want a UUID v7", got.ID)
		}
	})

	t.Run("Returns an opaque random token", func(t *testing.T) {
		t.Parallel()

		_, plain, err := NewRefreshToken(familyID, userID, expiresAt)
		if err != nil {
			t.Fatalf("NewRefreshToken() unexpected error: %v", err)
		}

		// Vai em JSON e em URL: base64 URL-safe sem padding.
		raw, err := base64.RawURLEncoding.DecodeString(plain)
		if err != nil {
			t.Fatalf("token %q is not base64 URL-safe without padding: %v", plain, err)
		}
		if len(raw) < 32 {
			t.Errorf("token has %d random bytes, want at least 32", len(raw))
		}
	})

	t.Run("Stores only the hash of the token", func(t *testing.T) {
		t.Parallel()

		got, plain, err := NewRefreshToken(familyID, userID, expiresAt)
		if err != nil {
			t.Fatalf("NewRefreshToken() unexpected error: %v", err)
		}

		if got.TokenHash != HashRefreshToken(plain) {
			t.Errorf("token_hash = %q, want HashRefreshToken(token) = %q", got.TokenHash, HashRefreshToken(plain))
		}
		if got.TokenHash == plain {
			t.Error("token_hash equals the plain token, want only the hash stored")
		}
	})

	t.Run("Generates a new ID and token on each call", func(t *testing.T) {
		t.Parallel()

		first, firstPlain, err := NewRefreshToken(familyID, userID, expiresAt)
		if err != nil {
			t.Fatalf("NewRefreshToken() unexpected error: %v", err)
		}
		second, secondPlain, err := NewRefreshToken(familyID, userID, expiresAt)
		if err != nil {
			t.Fatalf("NewRefreshToken() unexpected error: %v", err)
		}

		if first.ID == second.ID {
			t.Errorf("both calls returned id %q, want different IDs", first.ID)
		}
		if firstPlain == secondPlain {
			t.Errorf("both calls returned token %q, want different tokens", firstPlain)
		}
	})
}

func TestNewRefreshTokenValidation(t *testing.T) {
	t.Parallel()

	const (
		validFamilyID = "01a11d00-8e01-736d-9c09-a76d396bbc2f"
		validUserID   = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
		// UUID válido, mas versão 4.
		uuidV4 = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	)
	expiresAt := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		familyID string
		userID   string
		wantErr  []error
		notErr   []error
	}{
		{
			name:     "family id is not v7",
			familyID: uuidV4,
			userID:   validUserID,
			wantErr:  []error{ErrInvalidFamilyID},
			notErr:   []error{ErrInvalidUserID},
		},
		{
			name:     "user id is not a UUID",
			familyID: validFamilyID,
			userID:   "abc",
			wantErr:  []error{ErrInvalidUserID},
			notErr:   []error{ErrInvalidFamilyID},
		},
		{
			name:     "empty values",
			familyID: "",
			userID:   "",
			wantErr:  []error{ErrInvalidFamilyID, ErrInvalidUserID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, plain, err := NewRefreshToken(tt.familyID, tt.userID, expiresAt)
			if err == nil {
				t.Fatal("NewRefreshToken() = error nil; want error")
			}

			for _, want := range tt.wantErr {
				if !errors.Is(err, want) {
					t.Errorf("NewRefreshToken() = %v; want it to include %v", err, want)
				}
			}
			for _, not := range tt.notErr {
				if errors.Is(err, not) {
					t.Errorf("NewRefreshToken() = %v; want it not to include %v", err, not)
				}
			}

			if got != (RefreshToken{}) {
				t.Errorf("NewRefreshToken() = %+v; want zero RefreshToken on error", got)
			}
			if plain != "" {
				t.Errorf("NewRefreshToken() token = %q; want empty on error", plain)
			}
		})
	}
}

func TestHashRefreshToken(t *testing.T) {
	t.Parallel()

	t.Run("Returns the SHA-256 in hexadecimal", func(t *testing.T) {
		t.Parallel()

		// SHA-256 conhecido de "test".
		const want = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"

		if got := HashRefreshToken("test"); got != want {
			t.Errorf("HashRefreshToken(%q) = %q, want %q", "test", got, want)
		}
	})

	t.Run("Same token gives the same hash", func(t *testing.T) {
		t.Parallel()

		// O /refresh acha o token no banco pelo hash: sem isso a busca não funciona.
		first := HashRefreshToken("some-token")
		second := HashRefreshToken("some-token")

		if first != second {
			t.Error("HashRefreshToken() returned different hashes for the same token")
		}
	})

	t.Run("Different tokens give different hashes", func(t *testing.T) {
		t.Parallel()

		if HashRefreshToken("token-a") == HashRefreshToken("token-b") {
			t.Error("HashRefreshToken() returned the same hash for different tokens")
		}
	})
}

func TestRefreshTokenIsExpired(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	rt := RefreshToken{ExpiresAt: expiresAt}

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{name: "before expiration", now: expiresAt.Add(-time.Second), want: false},
		// expires_at é o primeiro instante em que o token já não vale.
		{name: "exactly at expiration", now: expiresAt, want: true},
		{name: "after expiration", now: expiresAt.Add(time.Second), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := rt.IsExpired(tt.now); got != tt.want {
				t.Errorf("IsExpired(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

func TestRefreshTokenIsRevoked(t *testing.T) {
	t.Parallel()

	if (RefreshToken{}).IsRevoked() {
		t.Error("IsRevoked() = true with zero revoked_at, want false")
	}

	rt := RefreshToken{RevokedAt: time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)}
	if !rt.IsRevoked() {
		t.Error("IsRevoked() = false with revoked_at set, want true")
	}
}

func TestRefreshTokenIsRotated(t *testing.T) {
	t.Parallel()

	if (RefreshToken{}).IsRotated() {
		t.Error("IsRotated() = true with zero rotated_at, want false")
	}

	rt := RefreshToken{RotatedAt: time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)}
	if !rt.IsRotated() {
		t.Error("IsRotated() = false with rotated_at set, want true")
	}
}
