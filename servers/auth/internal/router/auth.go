package router

import (
	"auth/internal/handler"
	"auth/internal/repository"
	"auth/internal/usecase"
	"net/http"
)

func registerCreateAuth(mux *http.ServeMux, d Deps) {
	r := repository.NewAuthRepository(d.DB)
	uc := usecase.NewCreateAuthUseCase(r)
	h := handler.CreateAuth(uc)

	mux.HandleFunc("POST /api/auth/create_user", h)
}

func RegisterAuth(mux *http.ServeMux, d Deps) {
	registerCreateAuth(mux, d)
}
