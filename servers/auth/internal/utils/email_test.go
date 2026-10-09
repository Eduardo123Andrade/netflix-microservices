package utils

import "testing"

func TestIsEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "simple", value: "teste@teste.com", want: true},
		{name: "subdomain", value: "teste@mail.teste.com.br", want: true},
		{name: "plus tag", value: "teste+netflix@teste.com", want: true},
		{name: "dots in local part", value: "nome.sobrenome@teste.com", want: true},
		{name: "uppercase", value: "Teste@Teste.com", want: true},

		{name: "empty", value: "", want: false},
		{name: "without at", value: "teste.teste.com", want: false},
		{name: "without local part", value: "@teste.com", want: false},
		{name: "without domain", value: "teste@", want: false},
		{name: "two at", value: "teste@@teste.com", want: false},
		{name: "space inside", value: "tes te@teste.com", want: false},
		{name: "leading space", value: " teste@teste.com", want: false},
		{name: "trailing space", value: "teste@teste.com ", want: false},
		{name: "with display name", value: "Teste <teste@teste.com>", want: false},
		{name: "only angle brackets", value: "<teste@teste.com>", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsEmail(tt.value); got != tt.want {
				t.Errorf("IsEmail(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
