package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dench1ka/vacancy-radar/internal/storage"
	"github.com/go-chi/chi/v5"
)

type createSubscriptionRequest struct {
	Keyword string `json:"keyword"`
	AreaID  string `json:"area_id"`
}

func (s *Server) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Keyword = strings.TrimSpace(req.Keyword)
	if req.Keyword == "" {
		writeError(w, http.StatusBadRequest, "keyword is required")
		return
	}

	sub, err := s.subscriptions.Create(r.Context(), userID, req.Keyword, req.AreaID)
	if err != nil {
		s.log.Error("create subscription failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not create subscription")
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

func (s *Server) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	subs, err := s.subscriptions.ListByUser(r.Context(), userID)
	if err != nil {
		s.log.Error("list subscriptions failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list subscriptions")
		return
	}

	writeJSON(w, http.StatusOK, subs)
}

func (s *Server) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	err := s.subscriptions.Delete(r.Context(), userID, id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "subscription not found")
		return
	}
	if err != nil {
		s.log.Error("delete subscription failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
