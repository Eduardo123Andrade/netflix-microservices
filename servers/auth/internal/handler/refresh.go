package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"auth/internal/usecase"
)

type refreshRotator interface {
	Rotate(ctx context.Context, refreshToken string) (userID, newRefreshToken string, err error)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh troca um refresh token por um par novo (rotação). Token
// desconhecido, vencido, revogado ou reusado fora da janela respondem igual
// (401); o reuso ainda vai para o log, porque indica token vazado.
func Refresh(rot refreshRotator, ts tokenService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req refreshRequest
		if !decodeJSON(w, r, &req) {
			return
		}

		userID, refresh, err := rot.Rotate(r.Context(), req.RefreshToken)
		switch {
		case err == nil:
		case errors.Is(err, usecase.ErrInvalidRefreshToken):
			if errors.Is(err, usecase.ErrRefreshTokenReused) {
				log.Printf("refresh: %v", err)
			}
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid refresh token"})
			return
		default:
			log.Printf("refresh: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
			return
		}

		// O refresh antigo já foi girado. Se o access falhar aqui, o cliente
		// tenta de novo com o antigo dentro da janela de tolerância.
		access, err := ts.Generate(userID)
		if err != nil {
			log.Printf("refresh: generate access token: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
			return
		}

		writeTokens(w, http.StatusOK, tokenResponse{AccessToken: access, TokenType: "Bearer", RefreshToken: refresh})
	}
}
