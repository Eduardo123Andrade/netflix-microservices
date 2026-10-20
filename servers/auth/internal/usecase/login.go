package usecase

import (
	"context"
	"errors"
	"fmt"

	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/services"
)

type hashService interface {
	passwordHasher
	passwordComparer
}

type LoginUseCase struct {
	repo     authFinder
	hs       hashService
	fakeHash string
}

type LoginData struct {
	Email    string
	Password string
}

func NewLoginUseCase(repo authFinder, hs hashService) (*LoginUseCase, error) {
	fakeHash, err := hs.Hash("fake_password")
	if err != nil {
		return nil, fmt.Errorf("login: generate dummy hash: %w", err)
	}
	return &LoginUseCase{repo: repo, hs: hs, fakeHash: fakeHash}, nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (l *LoginUseCase) Execute(ctx context.Context, data LoginData) (string, error) {
	normalizedEmail := entity.NormalizeEmail(data.Email)

	if err := entity.ValidateEmail(normalizedEmail); err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	auth, err := l.repo.FindByEmail(ctx, normalizedEmail)

	if errors.Is(err, repository.ErrAuthNotFound) {
		_ = l.hs.Compare(l.fakeHash, data.Password)

		return "", ErrInvalidCredentials
	}

	if err != nil {
		return "", fmt.Errorf("login: find by email: %w", err)
	}

	err = l.hs.Compare(auth.PasswordHash, data.Password)

	if errors.Is(err, services.ErrMismatchedHashAndPassword) {
		return "", ErrInvalidCredentials
	}

	if err != nil {
		return "", fmt.Errorf("login: compare hash: %w: %w", ErrInvalidCredentials, err)
	}

	return auth.UserID, nil
}
