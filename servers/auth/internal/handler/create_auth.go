package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/usecase"
)

type createAuthExecutor interface {
	Execute(ctx context.Context, data usecase.AuthData) (string, error)
}

type createAuthRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func CreateAuth(uc createAuthExecutor, ts tokenService, rt refreshIssuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createAuthRequest
		if !decodeJSON(w, r, &req) {
			return
		}

		userID, err := uc.Execute(r.Context(), usecase.AuthData{
			Name:     req.Name,
			Email:    req.Email,
			Password: req.Password,
		})

		switch {
		case err == nil:
			// A conta já foi criada: sem tokens, o cliente faz login em seguida.
			tokens, err := issueTokens(r.Context(), ts, rt, userID)
			if err != nil {
				log.Printf("create auth: %v", err)
				tokens = tokenResponse{TokenType: "Bearer"}
			}

			writeTokens(w, http.StatusCreated, tokens)

		case errors.Is(err, entity.ErrInvalidEmail), errors.Is(err, entity.ErrInvalidPassword):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		case errors.Is(err, repository.ErrAuthAlreadyExists):
			writeJSON(w, http.StatusConflict, errorResponse{Error: "email already registered"})
		default:
			log.Printf("create auth: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
	}
}
