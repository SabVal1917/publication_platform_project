package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubPinger struct {
	err   error
	calls int
}

func (p *stubPinger) PingContext(context.Context) error {
	p.calls++
	return p.err
}

func TestHealthHandlerReady(t *testing.T) {
	pinger := &stubPinger{}
	readiness := NewReadinessState()
	handler := NewHealthHandler(pinger, readiness)

	response := httptest.NewRecorder()
	handler.Ready(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("Ready() status = %d, want %d", response.Code, http.StatusOK)
	}
	if pinger.calls != 1 {
		t.Fatalf("PingContext() calls = %d, want 1", pinger.calls)
	}
}

func TestHealthHandlerReadyWhenDatabaseUnavailable(t *testing.T) {
	pinger := &stubPinger{err: errors.New("database unavailable")}
	handler := NewHealthHandler(pinger, NewReadinessState())

	response := httptest.NewRecorder()
	handler.Ready(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("Ready() status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestHealthHandlerDraining(t *testing.T) {
	pinger := &stubPinger{}
	readiness := NewReadinessState()
	handler := NewHealthHandler(pinger, readiness)
	readiness.StartDraining()

	readyResponse := httptest.NewRecorder()
	handler.Ready(readyResponse, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if readyResponse.Code != http.StatusServiceUnavailable {
		t.Fatalf("Ready() while draining status = %d, want %d", readyResponse.Code, http.StatusServiceUnavailable)
	}
	if pinger.calls != 0 {
		t.Fatalf("PingContext() calls while draining = %d, want 0", pinger.calls)
	}

	liveResponse := httptest.NewRecorder()
	handler.Live(liveResponse, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if liveResponse.Code != http.StatusOK {
		t.Fatalf("Live() while draining status = %d, want %d", liveResponse.Code, http.StatusOK)
	}
}
