// Package httpapi exposes the REST API: auth, subscription management and
// the feed of vacancies that were matched and sent to the current user.
package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/dench1ka/vacancy-radar/internal/auth"
	"github.com/dench1ka/vacancy-radar/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	tokens        *auth.TokenManager
	users         *storage.UserStore
	subscriptions *storage.SubscriptionStore
	vacancies     *storage.VacancyStore
	log           *slog.Logger
}

func NewServer(tokens *auth.TokenManager, users *storage.UserStore, subs *storage.SubscriptionStore, vacancies *storage.VacancyStore, log *slog.Logger) *Server {
	return &Server{tokens: tokens, users: users, subscriptions: subs, vacancies: vacancies, log: log}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(s.requestLogger)

	r.Get("/healthz", s.handleHealthz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", s.handleRegister)
		r.Post("/auth/login", s.handleLogin)

		r.Group(func(r chi.Router) {
			r.Use(s.requireAuth)

			r.Get("/me", s.handleGetMe)
			r.Put("/me/telegram", s.handleSetTelegramChatID)

			r.Post("/subscriptions", s.handleCreateSubscription)
			r.Get("/subscriptions", s.handleListSubscriptions)
			r.Delete("/subscriptions/{id}", s.handleDeleteSubscription)

			r.Get("/vacancies", s.handleListVacancies)
		})
	})

	return r
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
