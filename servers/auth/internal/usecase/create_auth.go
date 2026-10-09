package usecase

import (
	"auth/internal/entity"
	"auth/internal/repository"
	"context"
	"errors"
	"fmt"
)

type authRepository interface {
	CreateAuth(ctx context.Context, data entity.Auth) error
	FindByEmail(ctx context.Context, email string) (entity.Auth, error)
}

type hashService interface {
	Hash(pass string) (string, error)
}

type AuthData struct {
	Email, PasswordHash, UserID string
}

type AuthUseCase struct {
	repo authRepository
	hs   hashService
}

func NewCreateAuthUseCase(repo authRepository, hs hashService) *AuthUseCase {
	return &AuthUseCase{repo: repo, hs: hs}
}

func (a *AuthUseCase) Execute(ctx context.Context, data AuthData) error {
	authEntity, err := entity.NewAuth(data.Email, data.PasswordHash, data.UserID)

	if err != nil {
		return fmt.Errorf("create auth: %w", err)
	}

	passwordHash, err := a.hs.Hash(authEntity.PasswordHash)

	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	authEntity.PasswordHash = passwordHash

	_, err = a.repo.FindByEmail(ctx, data.Email)

	switch {
	case err == nil:
		return repository.ErrAuthAlreadyExists
	case !errors.Is(err, repository.ErrAuthNotFound):
		return fmt.Errorf("create auth: find by email: %w", err)
	}

	if err := a.repo.CreateAuth(ctx, authEntity); err != nil {
		return fmt.Errorf("create auth: save: %w", err)
	}

	return nil
}
