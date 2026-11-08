//go:build !sre_demo

package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func configureSREDemo(chi.Router) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc { return next }
}
