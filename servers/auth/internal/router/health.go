package router

import (
	"auth/internal/buildinfo"
	"auth/internal/handler"
	"auth/internal/repository"
	"auth/internal/usecase"
	"net/http"
)

func registerHealth(mux *http.ServeMux, d Deps) {
	r := repository.NewHealth(d.DB)
	uc := usecase.NewHealth(r, buildinfo.Version)
	hf := handler.Health(uc)
	mux.HandleFunc("GET /api/health", hf)
}
