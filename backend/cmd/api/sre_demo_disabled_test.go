//go:build !sre_demo

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestProductionHasNoSREDemoControl(t *testing.T) {
	router := chi.NewRouter()
	wrap := configureSREDemo(router)
	router.Get("/api/articles", wrap(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for _, mode := range []string{"normal", "error", "slow"} {
		control := httptest.NewRecorder()
		router.ServeHTTP(control, httptest.NewRequest(http.MethodPost, "/internal/sre-demo/"+mode, nil))
		if control.Code != http.StatusNotFound {
			t.Fatalf("production control status = %d, want 404", control.Code)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/articles", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("production response = %d, want 200", response.Code)
		}
	}
}
