package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/domain/models"
	"backend/internal/pkg/utils"
	"backend/internal/repository/postgres"
)

type UserHandler struct {
	Repo     *postgres.UserRepository
	TManager *utils.TokenManager
}

func NewUserHandler(repo *postgres.UserRepository, tokens *utils.TokenManager) *UserHandler {
	return &UserHandler{Repo: repo, TManager: tokens}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	hash, err := utils.HashPassword(u.PasswordHash)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка хеширования пароля")
		return
	}
	u.PasswordHash = hash

	if err := h.Repo.Create(&u); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка регистрации (возможно email занят)")
		return
	}

	token, err := h.TManager.GenerateToken(u.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}
	uResp := models.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  uResp,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	user, err := h.Repo.GetByEmail(input.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Пользователь с таким email не найден")
		return
	}

	if !utils.CheckPassword(input.Password, user.PasswordHash) {
		respondWithError(w, http.StatusUnauthorized, "Неверный пароль")
		return
	}

	token, err := h.TManager.GenerateToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}
	uResp := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Bio:       user.Bio,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  uResp,
	})
}
