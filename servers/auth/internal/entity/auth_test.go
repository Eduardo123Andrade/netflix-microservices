package entity

import (
	"auth/internal/utils"
	"testing"
)

func TestNewAuth(t *testing.T) {
	t.Parallel()

	const (
		email        = "teste@teste.com"
		passwordHash = "$2a$10$fakehash"
		userID       = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
	)

	t.Run("Fills fields from arguments", func(t *testing.T) {
		t.Parallel()

		got, err := NewAuth(email, passwordHash, userID)
		if err != nil {
			t.Fatalf("NewAuth() unexpected error: %v", err)
		}

		if got.Email != email {
			t.Errorf("email = %q, want %q", got.Email, email)
		}
		if got.PasswordHash != passwordHash {
			t.Errorf("password_hash = %q, want %q", got.PasswordHash, passwordHash)
		}
		if got.UserID != userID {
			t.Errorf("user_id = %q, want %q", got.UserID, userID)
		}
	})

	t.Run("Generates a UUID v7 ID", func(t *testing.T) {
		t.Parallel()

		got, err := NewAuth(email, passwordHash, userID)
		if err != nil {
			t.Fatalf("NewAuth() unexpected error: %v", err)
		}

		if !utils.IsUUIDV7(got.ID) {
			t.Errorf("id = %q, want a UUID v7", got.ID)
		}
	})

	t.Run("Generates a new ID on each call", func(t *testing.T) {
		t.Parallel()

		first, err := NewAuth(email, passwordHash, userID)
		if err != nil {
			t.Fatalf("NewAuth() unexpected error: %v", err)
		}
		second, err := NewAuth(email, passwordHash, userID)
		if err != nil {
			t.Fatalf("NewAuth() unexpected error: %v", err)
		}

		if first.ID == second.ID {
			t.Errorf("both calls returned id %q, want different IDs", first.ID)
		}
	})
}
