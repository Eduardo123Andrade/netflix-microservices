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

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "already normalized", value: "teste@teste.com", want: "teste@teste.com"},
		{name: "uppercase", value: "Teste@Teste.COM", want: "teste@teste.com"},
		{name: "leading and trailing spaces", value: "  teste@teste.com  ", want: "teste@teste.com"},
		{name: "tabs and newline around", value: "\tteste@teste.com\n", want: "teste@teste.com"},
		{name: "uppercase and spaces", value: " Teste@Teste.com ", want: "teste@teste.com"},
		{name: "empty", value: "", want: ""},
		{name: "only spaces", value: "   ", want: ""},
		// Normalizar não valida: o que não é email continua não sendo, só em minúsculas.
		{name: "not an email", value: " Teste.Teste.com ", want: "teste.teste.com"},
		{name: "display name is kept", value: "Teste <Teste@teste.com>", want: "teste <teste@teste.com>"},
		{name: "space inside is kept", value: "tes te@teste.com", want: "tes te@teste.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := NormalizeEmail(tt.value); got != tt.want {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// Normalizar e depois validar é a ordem que o entity usa. Aqui fica garantido
// que a normalização não "conserta" formatos que o IsEmail recusa.
func TestNormalizeThenValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "uppercase becomes valid", value: "Teste@Teste.com", want: true},
		{name: "spaces around become valid", value: "  teste@teste.com  ", want: true},
		{name: "display name stays invalid", value: "Teste <teste@teste.com>", want: false},
		{name: "only angle brackets stay invalid", value: "<teste@teste.com>", want: false},
		{name: "space inside stays invalid", value: "tes te@teste.com", want: false},
		{name: "without at stays invalid", value: " Teste.Teste.com ", want: false},
		{name: "only spaces stay invalid", value: "   ", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsEmail(NormalizeEmail(tt.value)); got != tt.want {
				t.Errorf("IsEmail(NormalizeEmail(%q)) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
