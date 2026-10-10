package router

import (
	"auth/internal/handler"
	"auth/internal/repository"
	"auth/internal/services"
	"auth/internal/usecase"
	"net/http"
)

func registerCreateAuth(mux *http.ServeMux, d Deps) {
	r := repository.NewAuthRepository(d.DB)

	hs := services.NewHasher(d.Cost)
	us := services.NewUserService(d.Users)

	uc := usecase.NewCreateAuthUseCase(r, hs, us)
	h := handler.CreateAuth(uc)

	mux.HandleFunc("POST /api/auth/create_user", h)
}

func RegisterAuth(mux *http.ServeMux, d Deps) {
	registerCreateAuth(mux, d)
}
