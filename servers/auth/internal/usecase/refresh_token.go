package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/utils"
)

type refreshTokenStore interface {
	refreshTokenCreator
	refreshTokenFinder
	refreshTokenRotationMarker
	refreshFamilyRevoker
}

type RefreshTokenUseCase struct {
	repo       refreshTokenStore
	refreshTTL time.Duration
	now        func() time.Time
}

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenReused  = errors.New("refresh token already used")
)

const (
	refreshGracePeriod = 5 * time.Minute
)

func NewRefreshTokenUseCase(repo refreshTokenStore, refreshTTL time.Duration, now func() time.Time) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		repo:       repo,
		refreshTTL: refreshTTL,
		now:        now,
	}
}

func (uc *RefreshTokenUseCase) Issue(ctx context.Context, userID string) (string, error) {
	familyID, err := utils.GenerateUUID()
	if err != nil {
		return "", fmt.Errorf("refresh token: issue: generate family id: %w", err)
	}

	plain, err := uc.create(ctx, familyID, userID, uc.now())
	if err != nil {
		return "", fmt.Errorf("refresh token: issue: %w", err)
	}

	return plain, nil
}

// create gera um token na família informada e o grava no store, devolvendo a
// chave pura para o cliente.
func (uc *RefreshTokenUseCase) create(ctx context.Context, familyID, userID string, now time.Time) (string, error) {
	rt, plain, err := entity.NewRefreshToken(familyID, userID, now.Add(uc.refreshTTL))
	if err != nil {
		return "", fmt.Errorf("new refresh token: %w", err)
	}

	if err := uc.repo.Create(ctx, rt); err != nil {
		return "", fmt.Errorf("save refresh token: %w", err)
	}

	return plain, nil
}

func (uc *RefreshTokenUseCase) Rotate(ctx context.Context, existingPlain string) (userID, plain string, err error) {
	if existingPlain == "" {
		return "", "", ErrInvalidRefreshToken
	}

	hash := entity.HashRefreshToken(existingPlain)

	rt, err := uc.repo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			return "", "", ErrInvalidRefreshToken
		}

		return "", "", fmt.Errorf("refresh token: rotate: find by hash: %w", err)
	}

	now := uc.now()

	if rt.IsExpired(now) || rt.IsRevoked() {
		return "", "", ErrInvalidRefreshToken
	}

	if rt.IsRotated() && now.Sub(rt.RotatedAt) > refreshGracePeriod {
		if err := uc.repo.RevokeFamily(ctx, rt.FamilyID, now); err != nil {
			return "", "", fmt.Errorf("refresh token: rotate: revoke family: %w", err)
		}
		return "", "", errors.Join(ErrInvalidRefreshToken, ErrRefreshTokenReused)
	}

	if err := uc.repo.MarkRotated(ctx, rt.ID, now); err != nil {
		return "", "", fmt.Errorf("refresh token: rotate: mark rotated: %w", err)
	}

	plain, err = uc.create(ctx, rt.FamilyID, rt.UserID, now)
	if err != nil {
		return "", "", fmt.Errorf("refresh token: rotate: %w", err)
	}

	return rt.UserID, plain, nil
}

func (uc *RefreshTokenUseCase) Revoke(ctx context.Context, plain string) error {
	if plain == "" {
		return nil
	}

	hash := entity.HashRefreshToken(plain)

	rt, err := uc.repo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			return nil
		}

		return fmt.Errorf("refresh token: revoke: find by hash: %w", err)
	}

	if err := uc.repo.RevokeFamily(ctx, rt.FamilyID, uc.now()); err != nil {
		return fmt.Errorf("refresh token: revoke: revoke family: %w", err)
	}

	return nil
}
