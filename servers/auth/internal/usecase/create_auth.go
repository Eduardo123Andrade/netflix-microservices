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

type AuthData struct {
	Email, PasswordHash, UserID string
}

type AuthUseCase struct {
	repo authRepository
}

func NewAuthUseCase(repo authRepository) *AuthUseCase {
	return &AuthUseCase{repo: repo}
}

func (a *AuthUseCase) Execute(ctx context.Context, data AuthData) error {
	authEntity, err := entity.NewAuth(data.Email, data.PasswordHash, data.UserID)

	if err != nil {
		return fmt.Errorf("create auth: %w", err)
	}

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
