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
func TestIntRange(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		min, max int
		unset    bool
		want     int
		wantErr  error
	}{
		{name: "int value inside range", envName: "KEY_1", min: 1, max: 10_000, envValue: "500", want: 500},
		{name: "value is text", envName: "KEY_2", min: 1, max: 10_000, envValue: "ABC", want: 0, wantErr: ErrNotInteger},
		{name: "value is float", envName: "KEY_3", min: 1, max: 10_000, envValue: "12.4", want: 0, wantErr: ErrNotInteger},
		{name: "value start with a integer", envName: "KEY_4", min: 1, max: 10_000, envValue: "12abc", want: 0, wantErr: ErrNotInteger},
		{name: "value end with a integer", envName: "KEY_5", min: 1, max: 10_000, envValue: "abc12", want: 0, wantErr: ErrNotInteger},
		{name: "value end have a integer", envName: "KEY_6", min: 1, max: 10_000, envValue: "abc12def", want: 0, wantErr: ErrNotInteger},
		{name: "value with space around", envName: "KEY_7", min: 1, max: 10_000, envValue: " 80 ", want: 80},
		{name: "value is negative", envName: "KEY_8", min: -10, max: 10_000, envValue: "-5", want: -5},
		{name: "value is lower than minimum", envName: "KEY_10", min: 1, max: 10_000, envValue: "0", want: 0, wantErr: ErrOutOfRange},
		{name: "value is the minimum value", envName: "KEY_11", min: 1, max: 10_000, envValue: "1", want: 1},
		{name: "value is higher than maximum", envName: "KEY_12", min: 1, max: 10_000, envValue: "10001", want: 0, wantErr: ErrOutOfRange},
		{name: "value is the maximum value", envName: "KEY_13", min: 1, max: 10_000, envValue: "10000", want: 10_000},
		{name: "value is empty", envName: "KEY_14", min: 1, max: 10_000, envValue: "", want: 0, wantErr: ErrEmpty},
		{name: "key doesnt exists", envName: "AUTH_TEST_NEVER_SET", min: 1, max: 10_000, want: 0, wantErr: ErrRequired, unset: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.unset {
				t.Setenv(tt.envName, tt.envValue)
			}

			r := Reader{}
			got := r.IntRange(tt.envName, tt.min, tt.max)

			if got != tt.want {
				t.Errorf("IntRange(%v) = %d; want %d", tt.envValue, got, tt.want)
			}

			if err := r.Err(); !errors.Is(err, tt.wantErr) {
				t.Errorf("IntRange(%q) error = %v; want %v", tt.envName, err, tt.wantErr)
			}

		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		unset    bool
		want     int
		wantErr  error
	}{
		{name: "positive value", envName: "KEY_INT_1", envValue: "500", want: 500},
		{name: "negative value", envName: "KEY_INT_2", envValue: "-5", want: -5},
		{name: "zero", envName: "KEY_INT_3", envValue: "0", want: 0},
		{name: "value with space around", envName: "KEY_INT_4", envValue: " 80 ", want: 80},
		{name: "value is text", envName: "KEY_INT_5", envValue: "ABC", want: 0, wantErr: ErrNotInteger},
		{name: "value is float", envName: "KEY_INT_6", envValue: "12.4", want: 0, wantErr: ErrNotInteger},
		{name: "value start with a integer", envName: "KEY_INT_7", envValue: "12abc", want: 0, wantErr: ErrNotInteger},
		{name: "value is empty", envName: "KEY_INT_8", envValue: "", want: 0, wantErr: ErrEmpty},
		{name: "key doesnt exists", envName: "AUTH_TEST_NEVER_SET", unset: true, want: 0, wantErr: ErrRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.unset {
				t.Setenv(tt.envName, tt.envValue)
			}

			r := Reader{}
			got := r.Int(tt.envName)

			if got != tt.want {
				t.Errorf("Int(%q) = %d; want %d", tt.envValue, got, tt.want)
			}

			if err := r.Err(); !errors.Is(err, tt.wantErr) {
				t.Errorf("Int(%q) error = %v; want %v", tt.envName, err, tt.wantErr)
			}
		})
	}
}
