package utils

import (
	"strings"
	"testing"
)

func TestIsValidPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "valid", value: "Senha@123", want: true},
		{name: "exactly 8 chars", value: "Abcdef!g", want: true},
		{name: "exactly 20 chars", value: "Abcdef!g" + strings.Repeat("x", 12), want: true},
		{name: "space counts as special", value: "Abc defg", want: true},

		{name: "empty", value: "", want: false},
		{name: "7 chars", value: "Abcde!g", want: false},
		{name: "21 chars", value: "Abcdef!g" + strings.Repeat("x", 13), want: false},
		{name: "without uppercase", value: "abcdef!g", want: false},
		{name: "without lowercase", value: "ABCDEF!G", want: false},
		{name: "without special", value: "Abcdefgh", want: false},
		{name: "digit is not special", value: "Abcdefg1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsValidPassword(tt.value); got != tt.want {
				t.Errorf("IsValidPassword(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
