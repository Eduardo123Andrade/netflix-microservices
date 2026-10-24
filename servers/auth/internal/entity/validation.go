package entity

import (
	"errors"

	"auth/internal/utils"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidUserID   = errors.New("invalid user id")
	ErrInvalidFamilyID = errors.New("invalid family id")
)

func ValidateEmail(email string) error {
	if !utils.IsEmail(email) {
		return ErrInvalidEmail
	}
	return nil
}

func ValidatePassword(password string) error {
	if !utils.IsValidPassword(password) {
		return ErrInvalidPassword
	}
	return nil
}

func ValidateUserID(userID string) error {
	if !utils.IsUUIDV7(userID) {
		return ErrInvalidUserID
	}
	return nil
}

func ValidateFamilyID(familyID string) error {
	if !utils.IsUUIDV7(familyID) {
		return ErrInvalidFamilyID
	}
	return nil
}
