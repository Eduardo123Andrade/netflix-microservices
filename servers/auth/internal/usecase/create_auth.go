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
	Hash(password string) (string, error)
}

type userService interface {
	CreateUser(ctx context.Context, name, email string) (userID string, err error)
	DeleteUser(ctx context.Context, userID string) error
}

type AuthData struct {
	Name, Email, Password string
}

type AuthUseCase struct {
	repo authRepository
	hs   hashService
	us   userService
}

func NewCreateAuthUseCase(repo authRepository, hs hashService, us userService) *AuthUseCase {
	return &AuthUseCase{repo: repo, hs: hs, us: us}
}

// Execute é a saga do cadastro. Tudo que pode falhar sem depender do serviço
// users vem antes do CreateUser; depois dele, qualquer falha desfaz o usuário.
func (a *AuthUseCase) Execute(ctx context.Context, data AuthData) error {
	if err := errors.Join(entity.ValidateEmail(data.Email), entity.ValidatePassword(data.Password)); err != nil {
		return fmt.Errorf("create auth: %w", err)
	}

	_, err := a.repo.FindByEmail(ctx, data.Email)
	switch {
	case err == nil:
		return repository.ErrAuthAlreadyExists
	case !errors.Is(err, repository.ErrAuthNotFound):
		return fmt.Errorf("create auth: find by email: %w", err)
	}

	passwordHash, err := a.hs.Hash(data.Password)
	if err != nil {
		return fmt.Errorf("create auth: hash password: %w", err)
	}

	userID, err := a.us.CreateUser(ctx, data.Name, data.Email)
	if err != nil {
		return fmt.Errorf("create auth: create user: %w", err)
	}

	authEntity, err := entity.NewAuth(data.Email, passwordHash, userID)
	if err != nil {
		return a.compensate(ctx, userID, fmt.Errorf("create auth: new auth: %w", err))
	}

	if err := a.repo.CreateAuth(ctx, authEntity); err != nil {
		return a.compensate(ctx, userID, fmt.Errorf("create auth: save: %w", err))
	}

	return nil
}

// compensate desfaz o usuário criado e devolve o erro original. Se a
// compensação também falhar, os dois erros voltam juntos: o usuário ficou
// órfão no serviço users e isso precisa aparecer no log.
func (a *AuthUseCase) compensate(ctx context.Context, userID string, cause error) error {
	if err := a.us.DeleteUser(ctx, userID); err != nil {
		return errors.Join(cause, fmt.Errorf("create auth: compensate: delete user %s: %w", userID, err))
	}
	return cause
}
