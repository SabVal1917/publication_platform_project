package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/domain/models"
	"backend/internal/repository/postgres"

	"github.com/go-chi/chi/v5"
)

type ArticleHandler struct {
	Repo *postgres.ArticleRepository
}

func NewArticleHandler(repo *postgres.ArticleRepository) *ArticleHandler {
	return &ArticleHandler{Repo: repo}
}

func (h *ArticleHandler) GetAllArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := h.Repo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка получения статей: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	if articles == nil {
		w.Write([]byte("[]"))
		return
	}
	
	json.NewEncoder(w).Encode(articles)
}

func (h *ArticleHandler) GetArticleByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	article, err := h.Repo.GetByID(id)
	if err != nil {
		http.Error(w, "Статья не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(article)
}

func (h *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var a models.Article

	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if a.Title == "" || a.Content == "" || a.Slug == "" {
		http.Error(w, "Заголовок, slug и контент обязательны", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(int)
	a.AuthorID = userID

	id, err := h.Repo.Create(&a)
	if err != nil {
		http.Error(w, "Ошибка сохранения: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (h *ArticleHandler) LikeArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Article ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	newCount, isLiked, err := h.Repo.ToggleLike(articleID, userID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"likes_count": newCount,
		"is_liked":    isLiked,
	})
}