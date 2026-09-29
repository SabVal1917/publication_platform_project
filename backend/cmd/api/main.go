package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/handlers"
	customMiddleware "backend/internal/handlers/middleware"
	"backend/internal/pkg/utils"
	"backend/internal/repository/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg, loadErr := config.Load()
	if loadErr != nil {
		log.Fatalf("load config: %v", loadErr)
	}
	db, dbErr := config.ConnectDB(cfg.Database)
	if dbErr != nil {
		log.Fatalf("db conn: %v", dbErr)
	}
	defer db.Close()
	tokens := utils.NewTokenManager(cfg.JWTSecret)

	articleRepo := postgres.NewArticleRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	userRepo := postgres.NewUserRepository(db)

	articleHandler := handlers.NewArticleHandler(articleRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo)
	userHandler := handlers.NewUserHandler(userRepo, tokens)
	healthHandler := handlers.NewHealthHandler(db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
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

	r.Post("/api/register", userHandler.Register)
	r.Post("/api/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(tokens))

		r.Post("/api/articles", articleHandler.CreateArticle)
		r.Post("/api/articles/{id}/comments", commentHandler.CreateComment)
		r.Post("/api/articles/{id}/like", articleHandler.LikeArticle)
	})

	// graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: r,
	}
	serverErrChan := make(chan error, 1)
	go func() {
		log.Printf("HTTP server starting on %s", server.Addr)
		serverErrChan <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalCtx.Done():
		log.Printf("shutdown signal received")

	case err := <-serverErrChan:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
		return
	}
	const shutdownTimeout = 5 * time.Second
	log.Printf("shutting down HTTP server (timeout: %s)", shutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		return
	}

	err := <-serverErrChan
	if !errors.Is(err, http.ErrServerClosed) {
		log.Printf("HTTP server stopped with err: %v", err)
	}
	log.Printf("HTTP server stopped gracefully")
}
