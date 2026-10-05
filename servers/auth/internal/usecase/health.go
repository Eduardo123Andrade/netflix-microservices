package usecase

import (
	"context"
	"log"
	"time"
)

const healthCheckTimeout = 2 * time.Second

type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

type HealthReport struct {
	Healthy bool
	Checks  map[string]Status
}
type healthChecker interface {
	Check(ctx context.Context) error
}

type Health struct {
	db healthChecker
}

func NewHealth(db healthChecker) *Health {
	return &Health{db: db}
}

func (h *Health) Execute(ctx context.Context) HealthReport {
	ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	report := HealthReport{
		Healthy: true,
		Checks:  map[string]Status{"database": StatusUp},
	}

	if err := h.db.Check(ctx); err != nil {
		log.Printf("health: database check failed: %v", err)
		report.Healthy = false
		report.Checks["database"] = StatusDown
	}

	return report
}
