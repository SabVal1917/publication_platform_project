package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/handlers"
	customMiddleware "backend/internal/handlers/middleware"
	"backend/internal/observability"
	"backend/internal/pkg/utils"
	"backend/internal/repository/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func main() {
	logger := observability.NewJSONLogger(os.Stdout)
	slog.SetDefault(logger)

	cfg, loadErr := config.Load()
	if loadErr != nil {
		logger.Error("load_config_failed", "error", loadErr)
		os.Exit(1)
	}
	dbConnectCtx, dbConnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbConnectCancel()
	db, dbErr := config.ConnectDB(dbConnectCtx, cfg.Database)
	if dbErr != nil {
		logger.Error("database_connection_failed", "error", dbErr)
		os.Exit(1)
	}
	defer db.Close()
	tokens := utils.NewTokenManager(cfg.JWTSecret)

	articleRepo := postgres.NewArticleRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	userRepo := postgres.NewUserRepository(db)

	articleHandler := handlers.NewArticleHandler(articleRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo)
	userHandler := handlers.NewUserHandler(userRepo, tokens)
	readiness := handlers.NewReadinessState()
	healthHandler := handlers.NewHealthHandler(db, readiness)
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	metrics := observability.NewMetrics(registry, db)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(observability.RequestLogger(logger))
	r.Use(metrics.Middleware)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Get("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Pong!"))
	})

	r.Get("/api/articles", articleHandler.GetAllArticles)
	r.Get("/api/articles/{id}", articleHandler.GetArticleByID)
	r.Get("/api/articles/{id}/comments", commentHandler.GetCommentsByArticle)
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)
	r.Handle("/metrics", metrics.Handler())

	r.Post("/api/register", userHandler.Register)
	r.Post("/api/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(tokens))

		r.Post("/api/articles", articleHandler.CreateArticle)
		r.Post("/api/articles/{id}/comments", commentHandler.CreateComment)
		r.Post("/api/articles/{id}/like", articleHandler.LikeArticle)
		r.Put("/api/articles/{id}", articleHandler.UpdateArticle)
		r.Delete("/api/articles/{id}", articleHandler.DeleteArticle)
	})

	// graceful shutdown
	server := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           r,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}
	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info("http_server_starting", "address", server.Addr)
		serverErrChan <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalCtx.Done():
		logger.Info("shutdown_signal_received")
		readiness.StartDraining()
		logger.Info("readiness_disabled", "drain_delay", cfg.HTTP.DrainDelay)
		if cfg.HTTP.DrainDelay > 0 {
			timer := time.NewTimer(cfg.HTTP.DrainDelay)
			select {
			case <-timer.C:
			case err := <-serverErrChan:
				timer.Stop()
				if !errors.Is(err, http.ErrServerClosed) {
					logger.Error("http_server_failed_while_draining", "error", err)
				}
				return
			}
		}

	case err := <-serverErrChan:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http_server_failed", "error", err)
		}
		return
	}
	logger.Info("http_server_shutting_down", "timeout", cfg.HTTP.ShutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("http_server_shutdown_failed", "error", err)
		if closeErr := server.Close(); closeErr != nil {
			logger.Error("http_server_force_close_failed", "error", closeErr)
		}
		return
	}

	err := <-serverErrChan
	if !errors.Is(err, http.ErrServerClosed) {
		logger.Error("http_server_stopped_with_error", "error", err)
	}
	logger.Info("http_server_stopped_gracefully")
}
