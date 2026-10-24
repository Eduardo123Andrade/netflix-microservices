package router

import (
	"fmt"
	"net/http"
	"time"

	"auth/internal/handler"
	"auth/internal/repository"
	"auth/internal/services"
	"auth/internal/usecase"
)

// RegisterAuth registra cadastro, login, refresh e logout. O usecase de
// refresh token é um só, compartilhado: cadastro e login abrem famílias,
// refresh gira e logout revoga.
func RegisterAuth(mux *http.ServeMux, d Deps) error {
	authRepo := repository.NewAuthRepository(d.DB)
	hasher := services.NewHasher(d.Cost)
	refresh := usecase.NewRefreshTokenUseCase(repository.NewRefreshTokenRepository(d.DB), d.RefreshTTL, time.Now)

	createAuth := usecase.NewCreateAuthUseCase(authRepo, hasher, services.NewUserService(d.Users))
	mux.HandleFunc("POST /api/auth/create_user", handler.CreateAuth(createAuth, d.Tokens, refresh))

	// O login pode falhar na montagem: ele gera o hash falso da proteção contra
	// ataque por tempo, e sem ele o serviço não deve subir.
	login, err := usecase.NewLoginUseCase(authRepo, hasher)
	if err != nil {
		return fmt.Errorf("register login: %w", err)
	}
	mux.HandleFunc("POST /api/auth/login", handler.Login(login, d.Tokens, refresh))

	mux.HandleFunc("POST /api/auth/refresh", handler.Refresh(refresh, d.Tokens))
	mux.HandleFunc("POST /api/auth/logout", handler.Logout(refresh))

	return nil
}
