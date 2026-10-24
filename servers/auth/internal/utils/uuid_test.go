package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateUUID(t *testing.T) {
	t.Parallel()

	t.Run("Returns a valid UUID v7", func(t *testing.T) {
		t.Parallel()

		got, err := GenerateUUID()
		if err != nil {
			t.Fatalf("GenerateUUID() unexpected error: %v", err)
		}

		parsed, err := uuid.Parse(got)
		if err != nil {
			t.Fatalf("GenerateUUID() = %q; want a valid UUID: %v", got, err)
		}
		if parsed.Version() != 7 {
			t.Errorf("version = %d, want 7", parsed.Version())
		}
		if parsed.Variant() != uuid.RFC4122 {
			t.Errorf("variant = %v, want %v", parsed.Variant(), uuid.RFC4122)
		}
		if parsed == uuid.Nil {
			t.Errorf("GenerateUUID() = %q; want a non-nil UUID", got)
		}
	})

	t.Run("Returns unique IDs", func(t *testing.T) {
		t.Parallel()

		const n = 1000
		seen := make(map[string]struct{}, n)

		for range n {
			got, err := GenerateUUID()
			if err != nil {
				t.Fatalf("GenerateUUID() unexpected error: %v", err)
			}
			if _, dup := seen[got]; dup {
				t.Fatalf("GenerateUUID() = %q; already generated", got)
			}
			seen[got] = struct{}{}
		}
	})

	t.Run("Returns IDs ordered by time", func(t *testing.T) {
		t.Parallel()

		const n = 1000
		prev, err := GenerateUUID()
		if err != nil {
			t.Fatalf("GenerateUUID() unexpected error: %v", err)
		}

		for range n {
			got, err := GenerateUUID()
			if err != nil {
				t.Fatalf("GenerateUUID() unexpected error: %v", err)
			}
			if got <= prev {
				t.Fatalf("GenerateUUID() = %q after %q; want increasing IDs", got, prev)
			}
			prev = got
		}
	})
}

func TestIsUUIDV7(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "valid v7", value: "01a11c10-8e01-736d-9c09-a76d396bbc2f", want: true},
		{name: "valid v4", value: "f47ac10b-58cc-4372-a567-0e02b2c3d479", want: false},
		{name: "nil UUID", value: "00000000-0000-0000-0000-000000000000", want: false},
		{name: "empty string", value: "", want: false},
		{name: "not a UUID", value: "abc", want: false},
		{name: "missing a character", value: "01a11c10-8e01-736d-9c09-a76d396bbc2", want: false},
		{name: "non hex character", value: "01a11c10-8e01-736d-9c09-a76d396bbc2g", want: false},

		// Formatos que o uuid.Parse aceita hoje. Se a regra passar a exigir
		// o formato padrão, troque o want destes casos para false.
		{name: "uppercase", value: "01A11C10-8E01-736D-9C09-A76D396BBC2F", want: true},
		{name: "with braces", value: "{01a11c10-8e01-736d-9c09-a76d396bbc2f}", want: true},
		{name: "with urn prefix", value: "urn:uuid:01a11c10-8e01-736d-9c09-a76d396bbc2f", want: true},
		{name: "without hyphens", value: "01a11c108e01736d9c09a76d396bbc2f", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsUUIDV7(tt.value); got != tt.want {
				t.Errorf("IsUUIDV7(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}

	t.Run("generated ID", func(t *testing.T) {
		t.Parallel()

		id, err := GenerateUUID()
		if err != nil {
			t.Fatalf("GenerateUUID() unexpected error: %v", err)
		}
		if !IsUUIDV7(id) {
			t.Errorf("IsUUIDV7(%q) = false, want true for a generated ID", id)
		}
	})
}
