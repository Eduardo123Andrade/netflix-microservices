package repository

import (
	"auth/internal/entity"
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

var (
	ErrAuthAlreadyExists = errors.New("duplicated data")
	ErrAuthNotFound      = errors.New("authentication not found")
)

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (ar *AuthRepository) CreateAuth(ctx context.Context, data entity.Auth) error {
	var pgError *pgconn.PgError

	_, err := ar.pool.Exec(ctx,
		"INSERT INTO auth (id, email, password_hash, user_id) VALUES($1, $2, $3, $4)",
		data.ID,
		data.Email,
		data.PasswordHash,
		data.UserID,
	)

	if errors.As(err, &pgError) {
		if pgError.Code == pgerrcode.UniqueViolation {
			return ErrAuthAlreadyExists
		}
	}

	return err
}

func (ar *AuthRepository) FindByEmail(ctx context.Context, email string) (entity.Auth, error) {
	var (
		gotID           string
		gotEmail        string
		gotPasswordHash string
		gotUserID       string
	)

	err := ar.pool.QueryRow(
		ctx,
		"SELECT a.id, a.email, a.password_hash, a.user_id FROM auth a WHERE a.email = $1",
		email,
	).Scan(&gotID, &gotEmail, &gotPasswordHash, &gotUserID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Auth{}, ErrAuthNotFound
		}
		return entity.Auth{}, err
	}

	return entity.Auth{
		ID:           gotID,
		Email:        gotEmail,
		PasswordHash: gotPasswordHash,
		UserID:       gotUserID,
	}, nil
}
