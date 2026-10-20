package handler

import (
	"context"
	"encoding/json"
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

// Login autentica por email e senha e devolve um access token. Email
// inexistente, senha errada e hash corrompido respondem igual (401), para o
// cliente não descobrir quais emails têm conta.
func Login(uc loginExecutor, ts tokenService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var req loginRequest
		if err := dec.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
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

		// Diferente do cadastro: no login o token é o próprio resultado, então
		// sem token não houve login.
		token, err := ts.Generate(userID)
		if err != nil {
			log.Printf("login: generate token: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
			return
		}

		writeJSON(w, http.StatusOK, tokenResponse{AccessToken: token, TokenType: "Bearer"})
	}
}
