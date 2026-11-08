package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func check(ctx context.Context, client *http.Client, target string) (bool, time.Duration) {
	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, time.Since(started)
	}
	response, err := client.Do(req)
	if err != nil {
		return false, time.Since(started)
	}
	defer response.Body.Close()
	_, err = io.Copy(io.Discard, response.Body)
	return err == nil && response.StatusCode == http.StatusOK, time.Since(started)
}

func main() {
	target := os.Getenv("PROBE_URL")
	if target == "" {
		target = "http://backend:8080/api/articles"
	}
	registry := prometheus.NewRegistry()
	success := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "publication_platform_probe_success",
		Help: "Whether the last HTTP probe completed with status 200 within two seconds.",
	})
	duration := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "publication_platform_probe_duration_seconds",
		Help: "Duration of the last HTTP probe including response body transfer.",
	})
	lastRun := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "publication_platform_probe_timestamp_seconds",
		Help: "Unix timestamp of the last completed HTTP probe.",
	})
	registry.MustRegister(success, duration, lastRun)
	client := &http.Client{
		Timeout: 2 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			ok, elapsed := check(ctx, client, target)
			value := 0.0
			if ok {
				value = 1
			}
			success.Set(value)
			duration.Set(elapsed.Seconds())
			lastRun.SetToCurrentTime()
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	server := &http.Server{Addr: ":8081", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("prober_server_failed", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("prober_shutdown_failed", "error", err)
	}
}
