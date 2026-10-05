package handler

import (
	"auth/internal/usecase"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type healthExecutor interface {
	Execute(ctx context.Context) usecase.HealthReport
}

type healthResponse struct {
	Status usecase.Status            `json:"status"`
	Checks map[string]usecase.Status `json:"checks"`
}

func Health(uc healthExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := uc.Execute(r.Context())

		status := http.StatusOK
		resp := healthResponse{Status: usecase.StatusUp, Checks: report.Checks}

		if !report.Healthy {
			status = http.StatusOK
			resp.Status = usecase.StatusUp
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("health: writing response: %v", err)
		}
	}
}
