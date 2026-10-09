package entity

import (
	"auth/internal/utils"
	"errors"
	"fmt"
)

type Auth struct {
	ID           string
	Email        string
	PasswordHash string
	UserID       string
}

var (
	ErrInvalidEmail  = errors.New("invalid email")
	ErrInvalidUserID = errors.New("invalid user id")
)

func NewAuth(email, passwordHash, userID string) (Auth, error) {
	var errs []error

	if !utils.IsEmail(email) {
		errs = append(errs, ErrInvalidEmail)
	}
	if !utils.IsUUIDV7(userID) {
		errs = append(errs, ErrInvalidUserID)
	}

	if len(errs) > 0 {
		return Auth{}, errors.Join(errs...)
	}

	id, err := utils.GenerateUUID()

	if err != nil {
		return Auth{}, fmt.Errorf("new auth: generate id: %w", err)
	}

	return Auth{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		UserID:       userID,
	}, nil
}
