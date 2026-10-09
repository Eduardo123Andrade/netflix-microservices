package services

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// Custo mínimo do bcrypt: o teste fica rápido e o comportamento é o mesmo.
const testCost = bcrypt.MinCost

func TestHashService(t *testing.T) {
	t.Parallel()

	const password = "Senha@123"

	t.Run("Hash matches the password", func(t *testing.T) {
		t.Parallel()

		hs := NewHasher(testCost)

		hash, err := hs.Hash(password)
		if err != nil {
			t.Fatalf("Hash() unexpected error: %v", err)
		}
		if hash == password {
			t.Fatal("Hash() returned the plain password")
		}

		if err := hs.Compare(hash, password); err != nil {
			t.Errorf("Compare() = %v; want nil for the right password", err)
		}
	})

	t.Run("Hash uses the configured cost", func(t *testing.T) {
		t.Parallel()

		hs := NewHasher(testCost)

		hash, err := hs.Hash(password)
		if err != nil {
			t.Fatalf("Hash() unexpected error: %v", err)
		}

		got, err := bcrypt.Cost([]byte(hash))
		if err != nil {
			t.Fatalf("bcrypt.Cost() unexpected error: %v", err)
		}
		if got != testCost {
			t.Errorf("cost = %d; want %d", got, testCost)
		}
	})

	t.Run("Same password gives different hashes", func(t *testing.T) {
		t.Parallel()

		hs := NewHasher(testCost)

		first, err := hs.Hash(password)
		if err != nil {
			t.Fatalf("Hash() unexpected error: %v", err)
		}
		second, err := hs.Hash(password)
		if err != nil {
			t.Fatalf("Hash() unexpected error: %v", err)
		}

		if first == second {
			t.Errorf("both hashes = %q; want different hashes (random salt)", first)
		}
	})

	t.Run("Error wrong password", func(t *testing.T) {
		t.Parallel()

		hs := NewHasher(testCost)

		hash, err := hs.Hash(password)
		if err != nil {
			t.Fatalf("Hash() unexpected error: %v", err)
		}

		err = hs.Compare(hash, "Senha@124")
		if !errors.Is(err, ErrMismatchedHashAndPassword) {
			t.Errorf("Compare() = %v; want %v", err, ErrMismatchedHashAndPassword)
		}
	})

	t.Run("Error corrupted hash is not a wrong password", func(t *testing.T) {
		t.Parallel()

		hs := NewHasher(testCost)

		err := hs.Compare("not-a-bcrypt-hash", password)
		if err == nil {
			t.Fatal("Compare() = nil; want error for a corrupted hash")
		}
		if errors.Is(err, ErrMismatchedHashAndPassword) {
			t.Errorf("Compare() = %v; want an error other than wrong password", err)
		}
	})

	t.Run("Error cost above bcrypt maximum", func(t *testing.T) {
		t.Parallel()

		hs := NewHasher(bcrypt.MaxCost + 1)

		if _, err := hs.Hash(password); err == nil {
			t.Error("Hash() = nil error; want error for cost above bcrypt maximum")
		}
	})

	t.Run("Error password above 72 bytes", func(t *testing.T) {
		t.Parallel()

		hs := NewHasher(testCost)

		if _, err := hs.Hash(strings.Repeat("a", 73)); err == nil {
			t.Error("Hash() = nil error; want error for password above 72 bytes")
		}
	})
}
