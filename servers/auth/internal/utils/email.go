package utils

import (
	"net/mail"
)

func IsEmail(email string) bool {
	address, err := mail.ParseAddress(email)

	if err != nil {
		return false
	}

	m := address.Address

	return m == email
}
