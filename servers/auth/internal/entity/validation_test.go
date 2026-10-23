package entity

import (
	"errors"
	"testing"
)

// As regras de formato em si são testadas em utils; aqui o que importa é
// cada validação devolver nil no caso válido e o erro sentinela no inválido.

const (
	validUUIDv7 = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"
	// UUID válido, mas versão 4.
	validUUIDv4 = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
)

type validationCase struct {
	name    string
	value   string
	wantErr error
}

func runValidationCases(t *testing.T, fnName string, validate func(string) error, tests []validationCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validate(tt.value)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("%s(%q) = %v; want nil", fnName, tt.value, err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("%s(%q) = %v; want %v", fnName, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	t.Parallel()

	runValidationCases(t, "ValidateEmail", ValidateEmail, []validationCase{
		{name: "valid", value: "teste@teste.com"},
		{name: "without @", value: "teste.teste.com", wantErr: ErrInvalidEmail},
		{name: "empty", value: "", wantErr: ErrInvalidEmail},
	})
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	runValidationCases(t, "ValidatePassword", ValidatePassword, []validationCase{
		{name: "valid", value: "Senha@123"},
		{name: "too short", value: "Se@1", wantErr: ErrInvalidPassword},
		{name: "without special character", value: "Senha1234", wantErr: ErrInvalidPassword},
		{name: "empty", value: "", wantErr: ErrInvalidPassword},
	})
}

func TestValidateUserID(t *testing.T) {
	t.Parallel()

	runValidationCases(t, "ValidateUserID", ValidateUserID, []validationCase{
		{name: "valid UUID v7", value: validUUIDv7},
		{name: "UUID v4", value: validUUIDv4, wantErr: ErrInvalidUserID},
		{name: "not a UUID", value: "abc", wantErr: ErrInvalidUserID},
		{name: "empty", value: "", wantErr: ErrInvalidUserID},
	})
}

func TestValidateFamilyID(t *testing.T) {
	t.Parallel()

	runValidationCases(t, "ValidateFamilyID", ValidateFamilyID, []validationCase{
		{name: "valid UUID v7", value: validUUIDv7},
		{name: "UUID v4", value: validUUIDv4, wantErr: ErrInvalidFamilyID},
		{name: "not a UUID", value: "abc", wantErr: ErrInvalidFamilyID},
		{name: "empty", value: "", wantErr: ErrInvalidFamilyID},
	})
}

func TestValidationErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	// O handler escolhe a resposta HTTP com errors.Is: um erro não pode ser
	// confundido com outro.
	errs := []error{ErrInvalidEmail, ErrInvalidPassword, ErrInvalidUserID, ErrInvalidFamilyID}

	for i, a := range errs {
		for _, b := range errs[i+1:] {
			if errors.Is(a, b) || errors.Is(b, a) {
				t.Errorf("%v and %v match each other, want distinct errors", a, b)
			}
		}
	}
}
