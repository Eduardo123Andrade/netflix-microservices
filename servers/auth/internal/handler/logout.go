package handler

import (
	"context"
	"log"
	"net/http"
)

type refreshRevoker interface {
	Revoke(ctx context.Context, refreshToken string) error
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout revoga a família do refresh token. O access token em uso não é
// revogado: ele vence sozinho em até JWT_TTL. Token desconhecido também
// responde 204, para o logout ser idempotente e não dizer se o token existe.
func Logout(rv refreshRevoker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req logoutRequest
		if !decodeJSON(w, r, &req) {
			return
		}

		if err := rv.Revoke(r.Context(), req.RefreshToken); err != nil {
			log.Printf("logout: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
