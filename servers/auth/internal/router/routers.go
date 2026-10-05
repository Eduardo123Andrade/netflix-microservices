package router

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	DB *pgxpool.Pool
}

func New(deps Deps) http.Handler {
	api := http.NewServeMux()

	registerHealth(api, deps)

	return api
}
