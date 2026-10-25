//go:build integration

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"backend/internal/handlers/middleware"
	"backend/internal/observability"
	"backend/internal/pkg/utils"
	"backend/internal/repository/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus"
)

func TestBusinessMetricsCountOnlySuccessfulOperations(t *testing.T) {
	db := businessMetricsTestDB(t)
	registry := prometheus.NewRegistry()
	metrics := observability.NewBusinessMetrics(registry)
	exporter := observability.NewMetrics(registry, db)
	tokens := utils.NewTokenManager("business-metrics-integration-test")
	users := NewUserHandler(postgres.NewUserRepository(db), tokens, metrics)
	articles := NewArticleHandler(postgres.NewArticleRepository(db), metrics)
	comments := NewCommentHandler(postgres.NewCommentRepository(db), metrics)

	router := chi.NewRouter()
	router.Post("/api/register", users.Register)
	router.Post("/api/login", users.Login)
	router.Handle("/metrics", exporter.Handler())
	router.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(tokens))
		r.Post("/api/articles", articles.CreateArticle)
		r.Post("/api/articles/{id}/comments", comments.CreateComment)
	})

	request := func(path, body, token string, wantStatus int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != wantStatus {
			t.Fatalf("POST %s: status = %d, want %d; body = %s", path, response.Code, wantStatus, response.Body.String())
		}
		return response
	}
	assertCounts := func(registered, logins, createdArticles, createdComments int) {
		t.Helper()
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		for name, value := range map[string]int{
			"users_registered_total": registered,
			"logins_total":           logins,
			"articles_created_total": createdArticles,
			"comments_created_total": createdComments,
		} {
			line := fmt.Sprintf("publication_platform_%s %d\n", name, value)
			if !bytes.Contains(response.Body.Bytes(), []byte(line)) {
				t.Fatalf("GET /metrics missing %q:\n%s", line, response.Body.String())
			}
		}
	}

	assertCounts(0, 0, 0, 0)
	openMetricsRequest := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	openMetricsRequest.Header.Set("Accept", "application/openmetrics-text;version=1.0.0")
	openMetricsResponse := httptest.NewRecorder()
	router.ServeHTTP(openMetricsResponse, openMetricsRequest)
	if !strings.HasPrefix(openMetricsResponse.Header().Get("Content-Type"), "application/openmetrics-text;") ||
		!strings.HasSuffix(openMetricsResponse.Body.String(), "# EOF\n") {
		t.Fatal("GET /metrics did not negotiate OpenMetrics")
	}
	signup := `{"username":"metrics-test","email":"metrics@example.com","password_hash":"test-password"}`
	request("/api/register", signup, "", http.StatusCreated)
	assertCounts(1, 0, 0, 0)
	request("/api/register", signup, "", http.StatusInternalServerError)
	assertCounts(1, 0, 0, 0)

	request("/api/login", `{"email":"metrics@example.com","password":"wrong-password"}`, "", http.StatusUnauthorized)
	request("/api/login", `{"email":"missing@example.com","password":"test-password"}`, "", http.StatusUnauthorized)
	assertCounts(1, 0, 0, 0)
	login := request("/api/login", `{"email":"metrics@example.com","password":"test-password"}`, "", http.StatusOK)
	var auth struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &auth); err != nil || auth.Token == "" {
		t.Fatalf("decode login token: %v", err)
	}
	assertCounts(1, 1, 0, 0)

	article := `{"title":"Metrics test","slug":"metrics-test","content":"Test article"}`
	request("/api/articles", article, "", http.StatusUnauthorized)
	request("/api/articles", `{}`, auth.Token, http.StatusBadRequest)
	assertCounts(1, 1, 0, 0)
	created := request("/api/articles", article, auth.Token, http.StatusCreated)
	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &result); err != nil || result.ID <= 0 {
		t.Fatalf("decode article ID: %v", err)
	}
	assertCounts(1, 1, 1, 0)
	// A duplicate slug rolls back the repository transaction.
	request("/api/articles", article, auth.Token, http.StatusInternalServerError)
	assertCounts(1, 1, 1, 0)

	path := fmt.Sprintf("/api/articles/%d/comments", result.ID)
	comment := `{"content":"Test comment"}`
	request(path, comment, "", http.StatusUnauthorized)
	request(path, `{`, auth.Token, http.StatusBadRequest)
	// An invalid foreign key rolls back the comment transaction.
	request("/api/articles/0/comments", comment, auth.Token, http.StatusInternalServerError)
	assertCounts(1, 1, 1, 0)
	request(path, comment, auth.Token, http.StatusCreated)
	assertCounts(1, 1, 1, 1)
}

// Each run uses a private schema so neither migrations nor fixtures touch app data.
func businessMetricsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid TEST_DATABASE_DSN")
	}
	admin := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = admin.Close() })
	schema := pgx.Identifier{fmt.Sprintf("business_metrics_test_%d", time.Now().UnixNano())}.Sanitize()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := admin.ExecContext(cleanupCtx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})
	config.RuntimeParams["search_path"] = schema
	db := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = db.Close() })
	for _, name := range []string{"000001_init.up.sql", "000002_consistent_counters_and_indexes.up.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		if err != nil {
			t.Fatalf("read migration: %v", err)
		}
		if _, err := db.ExecContext(ctx, string(migration)); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}
	return db
}
