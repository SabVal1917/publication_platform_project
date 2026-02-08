package main

import (
	"log"
	"net/http"

	"backend/internal/config"
	"backend/internal/handlers"
	customMiddleware "backend/internal/handlers/middleware"
	"backend/internal/repository/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	db := config.ConnectDB()
	defer db.Close()

	articleRepo := postgres.NewArticleRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	userRepo := postgres.NewUserRepository(db) 

	articleHandler := handlers.NewArticleHandler(articleRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo)
	userHandler := handlers.NewUserHandler(userRepo)
	
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, 
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Get("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Pong!"))
	})

	r.Get("/api/articles", articleHandler.GetAllArticles)
	r.Get("/api/articles/{id}", articleHandler.GetArticleByID)
	r.Get("/api/articles/{id}/comments", commentHandler.GetCommentsByArticle)
	
	r.Post("/api/register", userHandler.Register)
	r.Post("/api/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware)

		r.Post("/api/articles", articleHandler.CreateArticle)
		r.Post("/api/articles/{id}/comments", commentHandler.CreateComment)
		r.Post("/api/articles/{id}/like", articleHandler.LikeArticle)
	})



	log.Println("Server starting on port :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}