package env

import (
	"errors"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		unset    bool
		want     string
		wantErr  error
	}{
		{name: "valor simples", envName: "KEY_NAME", envValue: "TEST", want: "TEST"},
		{name: "espaços em volta", envName: "KEY_NAME", envValue: "  TEST  ", want: "TEST"},
		{name: "vazia", envName: "KEY_NAME", envValue: "", wantErr: ErrEmpty},
		{name: "só espaços", envName: "KEY_NAME", envValue: "   ", wantErr: ErrEmpty},
		{name: "não existe", envName: "AUTH_TEST_NEVER_SET", unset: true, wantErr: ErrRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.unset {
				t.Setenv(tt.envName, tt.envValue)
			}

			r := Reader{}
			got := r.String(tt.envName)

			if got != tt.want {
				t.Errorf("String(%q) = %q; want %q", tt.envName, got, tt.want)
			}

			if err := r.Err(); !errors.Is(err, tt.wantErr) {
				t.Errorf("String(%q) error = %v; want %v", tt.envName, err, tt.wantErr)
			}
		})
	}
}
