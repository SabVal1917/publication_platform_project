//go:build sre_demo

package main

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
)

func configureSREDemo(r chi.Router) func(http.HandlerFunc) http.HandlerFunc {
	var mode atomic.Int32
	r.Post("/internal/sre-demo/{mode}", func(w http.ResponseWriter, r *http.Request) {
		switch chi.URLParam(r, "mode") {
		case "normal":
			mode.Store(0)
		case "error":
			mode.Store(1)
		case "slow":
			mode.Store(2)
		default:
			http.Error(w, "expected normal, error or slow", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			switch mode.Load() {
			case 1:
				http.Error(w, "SRE demo failure", http.StatusInternalServerError)
				return
			case 2:
				timer := time.NewTimer(800 * time.Millisecond)
				defer timer.Stop()
				select {
				case <-timer.C:
				case <-r.Context().Done():
					return
				}
			}
			next(w, r)
		}
	}
}
