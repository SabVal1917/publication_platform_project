package handlers

import (
	"context"
	"net/http"
	"time"
)

type DatabasePinger interface {
	PingContext(context.Context) error
}

type HealthHandler struct {
	dbPinger DatabasePinger
}

func NewHealthHandler(dbPinger DatabasePinger) *HealthHandler {
	return &HealthHandler{dbPinger: dbPinger}
}

func (hh *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (hh *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*1)
	defer cancel()
	if err := hh.dbPinger.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
