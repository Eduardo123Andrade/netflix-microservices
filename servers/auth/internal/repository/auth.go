package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

type AuthData struct {
	ID           string
	Email        string
	PasswordHash string
	UserID       string
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (ar *AuthRepository) CreateAuth(ctx context.Context, data AuthData) error {
	_, err := ar.pool.Exec(ctx,
		"INSERT INTO auth (id, email, password_hash, user_id) VALUES($1, $2, $3, $4)",
		data.ID,
		data.Email,
		data.PasswordHash,
		data.UserID,
	)

	return err
}
