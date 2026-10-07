package database

import (
	"auth/internal/config"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNewPool = errors.New("unable to create database connections")
	ErrPing    = errors.New("unable to ping")
)

func NewPool(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:   cfg.Name,
	}

	connString := u.String()

	pool, err := pgxpool.New(ctx, connString)

	if err != nil {
		e := fmt.Errorf("%w: %w", ErrNewPool, err)
		return nil, e
	}

	return pool, nil
}

func Connect(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.Ping(ctx)

	if err != nil {
		e := fmt.Errorf("%w: %w", ErrPing, err)
		return e
	}

	return nil
}
