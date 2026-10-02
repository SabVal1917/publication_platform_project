package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"backend/internal/domain/models"
	"backend/internal/handlers/middleware"
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
	articles, err := h.Repo.GetAll(r.Context())
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

	article, err := h.Repo.GetByID(r.Context(), id)
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

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	a.AuthorID = userID

	id, err := h.Repo.Create(r.Context(), &a)
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

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	newCount, isLiked, err := h.Repo.ToggleLike(r.Context(), articleID, userID)
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

func (h *ArticleHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil || articleID <= 0 {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var a models.UpdateArticleRequest

	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(a.Title) == "" ||
		strings.TrimSpace(a.Content) == "" {
		respondWithError(w, http.StatusBadRequest, "Заголовок и содержимое обязательны")
		return
	}
	if utf8.RuneCountInString(a.Title) > 200 {
		respondWithError(w, http.StatusBadRequest, "Заголовок не должен превышать 200 символов")
		return
	}

	cleanTags := make([]string, 0, len(a.Tags))
	seenTags := make(map[string]struct{}, len(a.Tags))
	for _, rawTag := range a.Tags {
		tag := strings.TrimSpace(rawTag)
		if tag == "" {
			continue
		}
		if utf8.RuneCountInString(tag) > 50 {
			respondWithError(w, http.StatusBadRequest, "Тег не должен превышать 50 символов")
			return
		}

		tag = strings.ToLower(tag)
		if _, exists := seenTags[tag]; exists {
			continue
		}

		seenTags[tag] = struct{}{}
		cleanTags = append(cleanTags, tag)
	}
	a.Tags = cleanTags

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	err = h.Repo.Update(ctx, articleID, userID, a)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)

	case errors.Is(err, postgres.ErrArticleNotFound):
		respondWithError(w, http.StatusNotFound, "Статья не найдена")

	case errors.Is(err, postgres.ErrArticleForbidden):
		respondWithError(w, http.StatusForbidden, "Нельзя редактировать чужую статью")

	case errors.Is(err, context.DeadlineExceeded):
		respondWithError(w, http.StatusGatewayTimeout, "Превышено время ожидания базы данных")

	default:
		slog.ErrorContext(r.Context(), "update_article_failed", "article_id", articleID, "error", err)
		respondWithError(w, http.StatusInternalServerError, "Не удалось обновить статью")
	}
}

func (h *ArticleHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil || articleID <= 0 {
		respondWithError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.Repo.Delete(ctx, articleID, userID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, postgres.ErrArticleNotFound):
		respondWithError(w, http.StatusNotFound, "Статья не найдена")
	case errors.Is(err, postgres.ErrArticleForbidden):
		respondWithError(w, http.StatusForbidden, "Нельзя удалить чужую статью")
	case errors.Is(err, context.DeadlineExceeded):
		respondWithError(w, http.StatusGatewayTimeout, "Превышено время ожидания базы данных")
	default:
		slog.ErrorContext(r.Context(), "delete_article_failed", "article_id", articleID, "error", err)
		respondWithError(w, http.StatusInternalServerError, "Не удалось удалить статью")
	}
}
