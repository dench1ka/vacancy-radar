package httpapi

import (
	"net/http"
	"strconv"
)

func (s *Server) handleListVacancies(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	limit := parseIntDefault(r.URL.Query().Get("limit"), 20)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)

	vacancies, err := s.vacancies.ListForUser(r.Context(), userID, limit, offset)
	if err != nil {
		s.log.Error("list vacancies failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not list vacancies")
		return
	}

	writeJSON(w, http.StatusOK, vacancies)
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return def
	}
	return v
}
