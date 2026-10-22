package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const maxBodyBytes = 1 << 20 // 1 MB

type tokenService interface {
	Generate(userID string) (string, error)
}

type refreshIssuer interface {
	Issue(ctx context.Context, userID string) (string, error)
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// decodeJSON lê o corpo em dst recusando campos desconhecidos e corpos acima
// de maxBodyBytes. Se falhar, já responde 400 e devolve false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return false
	}
	return true
}

// issueTokens emite o par de uma sessão nova (cadastro e login): o access
// token primeiro, porque não grava nada; o refresh por último, porque abre
// uma família no banco.
func issueTokens(ctx context.Context, ts tokenService, rt refreshIssuer, userID string) (tokenResponse, error) {
	access, err := ts.Generate(userID)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("generate access token: %w", err)
	}

	refresh, err := rt.Issue(ctx, userID)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("issue refresh token: %w", err)
	}

	return tokenResponse{AccessToken: access, TokenType: "Bearer", RefreshToken: refresh}, nil
}

// writeTokens responde com tokens. O no-store impede proxies e o navegador
// de guardarem a resposta (RFC 6749 §5.1).
func writeTokens(w http.ResponseWriter, status int, tokens tokenResponse) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, status, tokens)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("writing response: %v", err)
	}
}
