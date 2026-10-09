package utils

import (
	"github.com/google/uuid"
)

func GenerateUUID() (string, error) {
	r, err := uuid.NewV7()

	if err != nil {
		return "", err
	}

	return r.String(), nil
}

func IsUUIDV7(value string) bool {
	u, err := uuid.Parse(value)

	if err != nil {
		return false
	}

	return u.Version() == 7
}
