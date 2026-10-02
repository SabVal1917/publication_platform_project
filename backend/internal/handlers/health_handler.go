package handlers

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"
)

type DatabasePinger interface {
	PingContext(context.Context) error
}

type HealthHandler struct {
	dbPinger DatabasePinger
	state    *ReadinessState
}

type ReadinessState struct {
	draining atomic.Bool
}

func NewReadinessState() *ReadinessState {
	return &ReadinessState{}
}

func (s *ReadinessState) StartDraining() {
	s.draining.Store(true)
}

func (s *ReadinessState) IsDraining() bool {
	return s.draining.Load()
}

func NewHealthHandler(dbPinger DatabasePinger, state *ReadinessState) *HealthHandler {
	return &HealthHandler{dbPinger: dbPinger, state: state}
}

func (hh *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (hh *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if hh.state.IsDraining() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*1)
	defer cancel()
	if err := hh.dbPinger.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
