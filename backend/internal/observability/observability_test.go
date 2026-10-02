package observability

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus"
)

func TestRequestLoggerWritesStructuredRecord(t *testing.T) {
	var output bytes.Buffer
	logger := NewJSONLogger(&output)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(RequestLogger(logger))
	router.Get("/articles/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/articles/42", nil))

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode JSON log: %v; output=%q", err, output.String())
	}
	if record["msg"] != "http_request" {
		t.Errorf("msg = %v, want http_request", record["msg"])
	}
	if record["method"] != http.MethodGet {
		t.Errorf("method = %v, want GET", record["method"])
	}
	if record["route"] != "/articles/{id}" {
		t.Errorf("route = %v, want /articles/{id}", record["route"])
	}
	if record["status"] != float64(http.StatusCreated) {
		t.Errorf("status = %v, want %d", record["status"], http.StatusCreated)
	}
	if record["request_id"] == "" {
		t.Error("request_id is empty")
	}
}

func TestRequestLoggerSkipsSuccessfulHealthCheck(t *testing.T) {
	var output bytes.Buffer
	logger := NewJSONLogger(&output)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if output.Len() != 0 {
		t.Fatalf("successful health check log = %q, want no log", output.String())
	}
}

func TestRequestLoggerKeepsFailedHealthCheck(t *testing.T) {
	var output bytes.Buffer
	logger := NewJSONLogger(&output)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if output.Len() == 0 {
		t.Fatal("failed health check was not logged")
	}
}

func TestMetricsMiddlewareUsesRoutePattern(t *testing.T) {
	db, err := sql.Open("pgx", "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	registry := prometheus.NewRegistry()
	metrics := NewMetrics(registry, db)

	router := chi.NewRouter()
	router.Use(metrics.Middleware)
	router.Get("/articles/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/articles/42", nil))

	response := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := response.Body.String()
	want := `publication_platform_http_requests_total{method="GET",route="/articles/{id}",status="204"} 1`
	if !strings.Contains(body, want) {
		t.Fatalf("metrics output does not contain %q:\n%s", want, body)
	}
	if !strings.Contains(body, "publication_platform_db_connections_open") {
		t.Fatalf("metrics output does not contain DB pool metrics:\n%s", body)
	}
}
