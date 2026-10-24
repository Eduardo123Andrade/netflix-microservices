package usecase

import (
	"context"
	"errors"
	"testing"
)

type fakeChecker struct {
	err error
}

func (f fakeChecker) Check(ctx context.Context) error {
	return f.err
}

func TestHealthDatabase(t *testing.T) {
	tests := []struct {
		name        string
		fake        fakeChecker
		wantHealthy bool
		wantStatus  Status
	}{
		{
			name:        "database up",
			fake:        fakeChecker{err: nil},
			wantHealthy: true,
			wantStatus:  StatusUp,
		},
		{
			name:        "database down",
			fake:        fakeChecker{err: errors.New("connection refused")},
			wantHealthy: false,
			wantStatus:  StatusDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHealth(tt.fake, "1.0.0-rc.1")
			got := h.Execute(context.Background())

			if got.Healthy != tt.wantHealthy {
				t.Errorf("Healthy = %v; want %v", got.Healthy, tt.wantHealthy)
			}
			if status := got.Checks["database"]; status != tt.wantStatus {
				t.Errorf(`Checks["database"] = %v; want %v`, status, tt.wantStatus)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	expect := "1.0.0-rc.1"

	h := NewHealth(fakeChecker{err: nil}, expect)
	r := h.Execute(context.Background())

	if r.Version != expect {
		t.Errorf("erro versão: %v; wants = %v", r.Version, expect)
	}

}
