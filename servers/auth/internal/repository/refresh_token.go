package repository

import (
	"context"
	"errors"
	"time"

	"auth/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (rtr *RefreshTokenRepository) Create(ctx context.Context, data entity.RefreshToken) error {
	_, err := rtr.pool.Exec(ctx,
		"INSERT INTO refresh_tokens (id, family_id, user_id, token_hash, expires_at) VALUES($1, $2, $3, $4, $5)",
		data.ID,
		data.FamilyID,
		data.UserID,
		data.TokenHash,
		data.ExpiresAt,
	)

	return err
}

func convertToTime(tt *time.Time) time.Time {
	if tt == nil {
		return time.Time{}
	}

	return *tt
}

func (rtr *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (entity.RefreshToken, error) {
	var (
		gotID        string
		gotFamilyID  string
		gotUserID    string
		gotTokenHash string
		gotExpiresAt time.Time
		gotRotatedAt *time.Time
		gotRevokedAt *time.Time
	)

	err := rtr.pool.QueryRow(
		ctx,
		"SELECT rt.id, rt.family_id, rt.user_id, rt.token_hash, rt.expires_at, rt.rotated_at, rt.revoked_at FROM refresh_tokens rt WHERE rt.token_hash = $1",
		hash,
	).Scan(
		&gotID,
		&gotFamilyID,
		&gotUserID,
		&gotTokenHash,
		&gotExpiresAt,
		&gotRotatedAt,
		&gotRevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.RefreshToken{}, ErrRefreshTokenNotFound
		}
		return entity.RefreshToken{}, err
	}

	return entity.RefreshToken{
		ID:        gotID,
		FamilyID:  gotFamilyID,
		UserID:    gotUserID,
		TokenHash: gotTokenHash,
		ExpiresAt: gotExpiresAt,
		RotatedAt: convertToTime(gotRotatedAt),
		RevokedAt: convertToTime(gotRevokedAt),
	}, nil
}

func (rtr *RefreshTokenRepository) MarkRotated(ctx context.Context, refreshID string, at time.Time) error {
	_, err := rtr.pool.Exec(ctx,
		"UPDATE refresh_tokens SET rotated_at = $1 WHERE id = $2 AND rotated_at IS NULL",
		at,
		refreshID,
	)

	return err
}

func (rtr *RefreshTokenRepository) RevokeFamily(ctx context.Context, familyID string, at time.Time) error {
	_, err := rtr.pool.Exec(ctx,
		"UPDATE refresh_tokens SET revoked_at = $1 WHERE family_id = $2 AND revoked_at IS NULL",
		at,
		familyID,
	)

	return err
}
