package utils

import (
	"net/mail"
	"strings"
)

func IsEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	m := address.Address

	return m == email
}

func NormalizeEmail(email string) string {
	e := strings.ToLower(email)
	e = strings.TrimSpace(e)

	return e
}
