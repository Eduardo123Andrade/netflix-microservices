package entity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"auth/internal/utils"
)

type RefreshToken struct {
	ID        string
	FamilyID  string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RotatedAt time.Time
	RevokedAt time.Time
}

func NewRefreshToken(familyID, userID string, expiresAt time.Time) (RefreshToken, string, error) {
	if err := errors.Join(ValidateFamilyID(familyID), ValidateUserID(userID)); err != nil {
		return RefreshToken{}, "", err
	}

	id, err := utils.GenerateUUID()
	if err != nil {
		return RefreshToken{}, "", fmt.Errorf("new refresh token: generate id: %w", err)
	}

	plain, err := generateRefreshSecret()
	if err != nil {
		return RefreshToken{}, "", fmt.Errorf("new refresh token: generate refresh secret: %w", err)
	}

	return RefreshToken{
		ID:        id,
		FamilyID:  familyID,
		UserID:    userID,
		TokenHash: HashRefreshToken(plain),
		ExpiresAt: expiresAt,
	}, plain, nil
}

func generateRefreshSecret() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	plain := base64.RawURLEncoding.EncodeToString(b)

	return plain, nil
}

func (rt RefreshToken) IsRevoked() bool {
	return !rt.RevokedAt.IsZero()
}

func (rt RefreshToken) IsRotated() bool {
	return !rt.RotatedAt.IsZero()
}

func (rt RefreshToken) IsExpired(now time.Time) bool {
	return !now.Before(rt.ExpiresAt)
}

func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
