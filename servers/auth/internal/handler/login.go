package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"auth/internal/entity"
	"auth/internal/usecase"
)

type loginExecutor interface {
	Execute(ctx context.Context, data usecase.LoginData) (string, error)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login autentica por email e senha e devolve access e refresh token. Email
// inexistente, senha errada e hash corrompido respondem igual (401), para o
// cliente não descobrir quais emails têm conta.
func Login(uc loginExecutor, ts tokenService, rt refreshIssuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if !decodeJSON(w, r, &req) {
			return
		}

		userID, err := uc.Execute(r.Context(), usecase.LoginData{
			Email:    req.Email,
			Password: req.Password,
		})

		switch {
		case err == nil:
		case errors.Is(err, entity.ErrInvalidEmail):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: entity.ErrInvalidEmail.Error()})
			return
		case errors.Is(err, usecase.ErrInvalidCredentials):
			// O erro "puro" é senha errada ou email inexistente; embrulhado, traz
			// uma causa do servidor (ex.: hash corrompido) que precisa ir para o log.
			if err != usecase.ErrInvalidCredentials {
				log.Printf("login: %v", err)
			}
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		default:
			log.Printf("login: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
			return
		}

		// Diferente do cadastro: no login os tokens são o próprio resultado, então
		// sem eles não houve login.
		tokens, err := issueTokens(r.Context(), ts, rt, userID)
		if err != nil {
			log.Printf("login: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
			return
		}

		writeTokens(w, http.StatusOK, tokens)
	}
}
