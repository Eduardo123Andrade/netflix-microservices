package usecase

import (
	"context"

	"auth/internal/entity"
)

type authFinder interface {
	FindByEmail(ctx context.Context, email string) (entity.Auth, error)
}

type authCreator interface {
	CreateAuth(ctx context.Context, data entity.Auth) error
}

type passwordHasher interface {
	Hash(password string) (string, error)
}

type passwordComparer interface {
	Compare(hash, password string) error
}

type userCreator interface {
	CreateUser(ctx context.Context, name, email string) (userID string, err error)
}

type userDeleter interface {
	DeleteUser(ctx context.Context, userID string) error
}
