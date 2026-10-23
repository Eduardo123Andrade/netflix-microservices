package entity

import (
	"errors"
	"fmt"

	"auth/internal/utils"
)

type Auth struct {
	ID           string
	Email        string
	PasswordHash string
	UserID       string
}

func NormalizeEmail(email string) string {
	return utils.NormalizeEmail(email)
}

func NewAuth(email, passwordHash, userID string) (Auth, error) {
	normalizedEmail := NormalizeEmail(email)

	if err := errors.Join(ValidateEmail(normalizedEmail), ValidateUserID(userID)); err != nil {
		return Auth{}, err
	}

	id, err := utils.GenerateUUID()
	if err != nil {
		return Auth{}, fmt.Errorf("new auth: generate id: %w", err)
	}

	return Auth{
		ID:           id,
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		UserID:       userID,
	}, nil
}
