package router

import (
	"auth/internal/handler/health"
	"net/http"
)

func New() http.Handler {
	api := http.NewServeMux()

	api.HandleFunc("GET /api/health", health.Health)

	return api
}
