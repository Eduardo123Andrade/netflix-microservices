package router

import (
	usersv1 "auth/internal/gen/users/v1"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	DB    *pgxpool.Pool
	Cost  int
	Users usersv1.UserServiceClient
}

func New(deps Deps) http.Handler {
	api := http.NewServeMux()

	registerHealth(api, deps)
	RegisterAuth(api, deps)

	return api
}
