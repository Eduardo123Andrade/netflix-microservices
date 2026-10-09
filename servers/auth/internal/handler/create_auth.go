package handler

import (
	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/usecase"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type createAuthExecutor interface {
	Execute(ctx context.Context, data usecase.AuthData) error
}

type authResponse struct {
	Message string `json:"message"`
}

const maxBodyBytes = 1 << 20 // 1 MB

type createAuthRequest struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	UserID       string `json:"user_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func CreateAuth(uc createAuthExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var req createAuthRequest
		if err := dec.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
			return
		}

		err := uc.Execute(r.Context(), usecase.AuthData{
			Email:        req.Email,
			PasswordHash: req.PasswordHash,
			UserID:       req.UserID,
		})

		switch {
		case err == nil:
			w.WriteHeader(http.StatusCreated)
		case errors.Is(err, entity.ErrInvalidEmail), errors.Is(err, entity.ErrInvalidUserID):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		case errors.Is(err, repository.ErrAuthAlreadyExists):
			writeJSON(w, http.StatusConflict, errorResponse{Error: "email already registered"})
		default:
			log.Printf("create auth: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}

	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("writing response: %v", err)
	}
}
