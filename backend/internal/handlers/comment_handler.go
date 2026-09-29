package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/domain/models"
	"backend/internal/handlers/middleware"
	"backend/internal/repository/postgres"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	Repo *postgres.CommentRepository
}

func NewCommentHandler(repo *postgres.CommentRepository) *CommentHandler {
	return &CommentHandler{Repo: repo}
}

func (h *CommentHandler) GetCommentsByArticle(w http.ResponseWriter, r *http.Request) {
	articleIDStr := chi.URLParam(r, "id")
	articleID, _ := strconv.Atoi(articleIDStr)

	comments, err := h.Repo.GetByArticleID(articleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if comments == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(comments)
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	articleIDStr := chi.URLParam(r, "id")
	articleID, _ := strconv.Atoi(articleIDStr)

	var c models.Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	c.ArticleID = articleID
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	c.AuthorID = userID

	if err := h.Repo.Create(&c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}
