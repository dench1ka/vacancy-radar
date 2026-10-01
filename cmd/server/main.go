// Command server runs the vacancy-radar HTTP API together with the background
// worker that polls hh.ru and sends Telegram notifications.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dench1ka/vacancy-radar/internal/auth"
	"github.com/dench1ka/vacancy-radar/internal/config"
	"github.com/dench1ka/vacancy-radar/internal/hhclient"
	"github.com/dench1ka/vacancy-radar/internal/httpapi"
	"github.com/dench1ka/vacancy-radar/internal/logger"
	"github.com/dench1ka/vacancy-radar/internal/storage"
	"github.com/dench1ka/vacancy-radar/internal/telegram"
	"github.com/dench1ka/vacancy-radar/internal/worker"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := storage.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	users := storage.NewUserStore(db)
	subscriptions := storage.NewSubscriptionStore(db)
	vacancies := storage.NewVacancyStore(db)

	tokens := auth.NewTokenManager(cfg.JWTSecret, 24*time.Hour)
	hh := hhclient.New(cfg.HHBaseURL)
	tg := telegram.New(cfg.TelegramToken)

	workerStore := storage.NewWorkerStore(subscriptions, vacancies)
	w := worker.New(workerStore, hh, tg, cfg.PollInterval, log)

	server := httpapi.NewServer(tokens, users, subscriptions, vacancies, log)
	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      server.Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go w.Run(ctx)

	go func() {
		log.Info("starting http server", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
}
