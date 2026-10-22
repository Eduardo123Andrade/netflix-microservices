package router

import (
	usersv1 "auth/internal/gen/users/v1"
	"auth/internal/services"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	DB     *pgxpool.Pool
	Cost   int
	Users  usersv1.UserServiceClient
	Tokens *services.TokenService
	// RefreshTTL é a validade de cada refresh token (JWT_REFRESH_TTL).
	RefreshTTL time.Duration
}

func New(deps Deps) (http.Handler, error) {
	api := http.NewServeMux()

	registerHealth(api, deps)
	if err := RegisterAuth(api, deps); err != nil {
		return nil, err
	}

	return api, nil
}
