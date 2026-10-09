package utils

import "unicode"

func IsValidPassword(value string) bool {
	if len(value) < 8 || len(value) > 20 {
		return false
	}

	var hasUpper, hasLower, hasSpecial bool

	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasSpecial
}
