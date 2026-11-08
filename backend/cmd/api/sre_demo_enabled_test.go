//go:build sre_demo

package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/internal/observability"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
)

func TestSREDemoModesPreserveRouteMetrics(t *testing.T) {
	db, err := sql.Open("pgx", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	metrics := observability.NewMetrics(prometheus.NewRegistry(), db)
	router := chi.NewRouter()
	router.Use(metrics.Middleware)
	wrap := configureSREDemo(router)
	router.Get("/api/articles", wrap(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	router.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for _, tc := range []struct {
		mode   string
		status int
	}{
		{"error", http.StatusInternalServerError},
		{"slow", http.StatusOK},
		{"normal", http.StatusOK},
	} {
		control := httptest.NewRecorder()
		router.ServeHTTP(control, httptest.NewRequest(http.MethodPost, "/internal/sre-demo/"+tc.mode, nil))
		if control.Code != http.StatusNoContent {
			t.Fatalf("control %s = %d", tc.mode, control.Code)
		}
		response := httptest.NewRecorder()
		started := time.Now()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/articles", nil))
		if response.Code != tc.status {
			t.Fatalf("mode %s status = %d, want %d", tc.mode, response.Code, tc.status)
		}
		if tc.mode == "slow" && time.Since(started) < 800*time.Millisecond {
			t.Fatal("slow request completed before the injected delay")
		}
		health := httptest.NewRecorder()
		router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health/live", nil))
		if health.Code != http.StatusOK {
			t.Fatalf("mode %s affected liveness", tc.mode)
		}
	}
	response := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	want := `publication_platform_http_requests_total{method="GET",route="/api/articles",status="500"} 1`
	if !strings.Contains(response.Body.String(), want) {
		t.Fatalf("missing route-labelled failure metric: %s", response.Body.String())
	}
	for _, want := range []string{
		`publication_platform_http_request_duration_seconds_bucket{method="GET",route="/api/articles",status="200",le="0.5"} 1`,
		`publication_platform_http_request_duration_seconds_count{method="GET",route="/api/articles",status="200"} 2`,
	} {
		if !strings.Contains(response.Body.String(), want) {
			t.Fatalf("missing latency metric %q: %s", want, response.Body.String())
		}
	}
}
