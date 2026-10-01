package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dench1ka/vacancy-radar/internal/auth"
	"github.com/dench1ka/vacancy-radar/internal/storage"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "email is required and password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not process password")
		return
	}

	user, err := s.users.Create(r.Context(), req.Email, hash)
	if errors.Is(err, storage.ErrAlreadyExists) {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		s.log.Error("create user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	token, err := s.tokens.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{Token: token})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.users.GetByEmail(r.Context(), req.Email)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		s.log.Error("get user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not log in")
		return
	}
	if errors.Is(err, storage.ErrNotFound) || !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := s.tokens.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token})
}

func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	user, err := s.users.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type setTelegramRequest struct {
	TelegramChatID int64 `json:"telegram_chat_id"`
}

func (s *Server) handleSetTelegramChatID(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	var req setTelegramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.users.SetTelegramChatID(r.Context(), userID, req.TelegramChatID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update telegram chat id")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
