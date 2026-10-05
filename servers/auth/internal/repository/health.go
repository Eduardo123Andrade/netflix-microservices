package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Health struct {
	pool *pgxpool.Pool
}

func NewHealth(pool *pgxpool.Pool) *Health {
	return &Health{pool: pool}
}

func (h *Health) Check(ctx context.Context) error {
	return h.pool.Ping(ctx)
}
