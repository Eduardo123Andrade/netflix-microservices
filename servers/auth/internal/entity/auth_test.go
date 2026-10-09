package entity

import (
	"auth/internal/utils"
	"errors"
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

func TestNewAuthValidation(t *testing.T) {
	t.Parallel()

	const (
		validEmail   = "teste@teste.com"
		passwordHash = "$2a$10$fakehash"
		validUserID  = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
		// UUID válido, mas versão 4.
		userIDv4 = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	)

	tests := []struct {
		name    string
		email   string
		userID  string
		wantErr []error
		notErr  []error
	}{
		{
			name:    "invalid email",
			email:   "teste.teste.com",
			userID:  validUserID,
			wantErr: []error{ErrInvalidEmail},
			notErr:  []error{ErrInvalidUserID},
		},
		{
			name:    "user id is not v7",
			email:   validEmail,
			userID:  userIDv4,
			wantErr: []error{ErrInvalidUserID},
			notErr:  []error{ErrInvalidEmail},
		},
		{
			name:    "user id is not a UUID",
			email:   validEmail,
			userID:  "abc",
			wantErr: []error{ErrInvalidUserID},
			notErr:  []error{ErrInvalidEmail},
		},
		{
			name:    "invalid email and user id",
			email:   "teste.teste.com",
			userID:  "abc",
			wantErr: []error{ErrInvalidEmail, ErrInvalidUserID},
		},
		{
			name:    "empty values",
			email:   "",
			userID:  "",
			wantErr: []error{ErrInvalidEmail, ErrInvalidUserID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewAuth(tt.email, passwordHash, tt.userID)
			if err == nil {
				t.Fatal("NewAuth() = error nil; want error")
			}

			for _, want := range tt.wantErr {
				if !errors.Is(err, want) {
					t.Errorf("NewAuth() = %v; want it to include %v", err, want)
				}
			}
			for _, not := range tt.notErr {
				if errors.Is(err, not) {
					t.Errorf("NewAuth() = %v; want it not to include %v", err, not)
				}
			}

			if got != (Auth{}) {
				t.Errorf("NewAuth() = %+v; want zero Auth on error", got)
			}
		})
	}
}
