package entity

import (
	"auth/internal/utils"
	"fmt"
)

type Auth struct {
	ID           string
	Email        string
	PasswordHash string
	UserID       string
}

func NewAuth(email, passwordHash, userID string) (Auth, error) {
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
