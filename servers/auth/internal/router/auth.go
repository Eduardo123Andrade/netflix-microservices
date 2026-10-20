package router

import (
	"fmt"
	"net/http"

	"auth/internal/handler"
	"auth/internal/repository"
	"auth/internal/services"
	"auth/internal/usecase"
)

func registerCreateAuth(mux *http.ServeMux, d Deps) {
	r := repository.NewAuthRepository(d.DB)

	hs := services.NewHasher(d.Cost)
	us := services.NewUserService(d.Users)

	uc := usecase.NewCreateAuthUseCase(r, hs, us)
	h := handler.CreateAuth(uc, d.Tokens)

	mux.HandleFunc("POST /api/auth/create_user", h)
}

// registerLogin pode falhar: o usecase gera o hash falso da proteção contra
// ataque por tempo na montagem, e sem ele o serviço não deve subir.
func registerLogin(mux *http.ServeMux, d Deps) error {
	r := repository.NewAuthRepository(d.DB)
	hs := services.NewHasher(d.Cost)

	uc, err := usecase.NewLoginUseCase(r, hs)
	if err != nil {
		return fmt.Errorf("register login: %w", err)
	}

	mux.HandleFunc("POST /api/auth/login", handler.Login(uc, d.Tokens))

	return nil
}

func RegisterAuth(mux *http.ServeMux, d Deps) error {
	registerCreateAuth(mux, d)

	return registerLogin(mux, d)
}
