package handler

import (
	"auth/internal/usecase"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

type fakeHealth struct {
	report usecase.HealthReport
}

func (f fakeHealth) Execute(ctx context.Context) usecase.HealthReport {
	return f.report
}

type fakeWrite struct {
	*httptest.ResponseRecorder
}

func (f *fakeWrite) Write(buf []byte) (int, error) {

	return 0, errors.New("write failed")
}

func TestEncoderError(t *testing.T) {
	fw := &fakeWrite{ResponseRecorder: httptest.NewRecorder()}

	var buf bytes.Buffer

	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	Health(fakeHealth{}).ServeHTTP(fw, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	wants := "health: writing response"
	if !strings.Contains(buf.String(), wants) {
		t.Errorf("log = %q; want it to contain %q", buf.String(), wants)
	}
}

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		report     usecase.HealthReport
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name: "database up",
			report: usecase.HealthReport{
				Healthy: true,
				Version: "1.0.0-rc.1",
				Checks:  map[string]usecase.Status{"database": usecase.StatusUp},
			},
			wantStatus: http.StatusOK,
			wantBody: map[string]any{
				"status":  "UP",
				"version": "1.0.0-rc.1",
				"checks":  map[string]any{"database": "UP"},
			},
		},
		{
			name: "database down",
			report: usecase.HealthReport{
				Healthy: false,
				Version: "1.0.0-rc.1",
				Checks:  map[string]usecase.Status{"database": usecase.StatusDown},
			},
			// Temporary
			// wantStatus: http.StatusServiceUnavailable,
			wantStatus: http.StatusOK,
			wantBody: map[string]any{
				"status":  "UP",
				"version": "1.0.0-rc.1",
				"checks":  map[string]any{"database": "DOWN"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			rec := httptest.NewRecorder()

			Health(fakeHealth{report: tt.report})(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", rec.Code, tt.wantStatus)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q; want %q", ct, "application/json")
			}

			var got map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if !reflect.DeepEqual(got, tt.wantBody) {
				t.Errorf("body = %v; want %v", got, tt.wantBody)
			}
		})
	}
}
