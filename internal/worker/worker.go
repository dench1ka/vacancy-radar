// Package worker periodically polls hh.ru on behalf of every active subscription
// and notifies users via Telegram about vacancies they haven't seen yet.
package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dench1ka/vacancy-radar/internal/hhclient"
)

// SubscriptionWithUser is a subscription joined with the data the worker needs
// from its owning user.
type SubscriptionWithUser struct {
	SubscriptionID string
	UserID         string
	Keyword        string
	AreaID         string
	TelegramChatID *int64
}

type Store interface {
	ListActiveWithUsers(ctx context.Context) ([]SubscriptionWithUser, error)
	IsSeen(ctx context.Context, userID, hhVacancyID string) (bool, error)
	MarkSeen(ctx context.Context, userID string, v hhclient.Vacancy) error
}

type HHSearcher interface {
	Search(ctx context.Context, p hhclient.SearchParams) ([]hhclient.Vacancy, error)
}

type Notifier interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type Worker struct {
	store    Store
	hh       HHSearcher
	notifier Notifier
	interval time.Duration
	log      *slog.Logger
}

func New(store Store, hh HHSearcher, notifier Notifier, interval time.Duration, log *slog.Logger) *Worker {
	if log == nil {
		log = slog.Default()
	}
	return &Worker{store: store, hh: hh, notifier: notifier, interval: interval, log: log}
}

// Run blocks, polling on a ticker until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.runOnceLogged(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnceLogged(ctx)
		}
	}
}

func (w *Worker) runOnceLogged(ctx context.Context) {
	if err := w.RunOnce(ctx); err != nil {
		w.log.Error("worker run failed", "error", err)
	}
}

// RunOnce polls hh.ru for every active subscription concurrently and notifies
// users about any vacancy they haven't been sent before.
func (w *Worker) RunOnce(ctx context.Context) error {
	subs, err := w.store.ListActiveWithUsers(ctx)
	if err != nil {
		return fmt.Errorf("list active subscriptions: %w", err)
	}

	const maxConcurrency = 5
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, sub := range subs {
		sub := sub
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := w.processSubscription(ctx, sub); err != nil {
				w.log.Error("process subscription failed", "subscription_id", sub.SubscriptionID, "error", err)
			}
		}()
	}

	wg.Wait()
	return nil
}

func (w *Worker) processSubscription(ctx context.Context, sub SubscriptionWithUser) error {
	vacancies, err := w.hh.Search(ctx, hhclient.SearchParams{Text: sub.Keyword, AreaID: sub.AreaID})
	if err != nil {
		return fmt.Errorf("search hh.ru: %w", err)
	}

	for _, v := range vacancies {
		seen, err := w.store.IsSeen(ctx, sub.UserID, v.HHVacancyID)
		if err != nil {
			w.log.Error("check seen failed", "error", err)
			continue
		}
		if seen {
			continue
		}

		if err := w.store.MarkSeen(ctx, sub.UserID, v); err != nil {
			w.log.Error("mark seen failed", "error", err)
			continue
		}

		if sub.TelegramChatID == nil {
			continue
		}
		if err := w.notifier.SendMessage(ctx, *sub.TelegramChatID, formatMessage(v)); err != nil {
			w.log.Error("send telegram message failed", "error", err)
		}
	}

	return nil
}

func formatMessage(v hhclient.Vacancy) string {
	salary := "не указана"
	if v.SalaryFrom != nil || v.SalaryTo != nil {
		switch {
		case v.SalaryFrom != nil && v.SalaryTo != nil:
			salary = fmt.Sprintf("%d–%d %s", *v.SalaryFrom, *v.SalaryTo, v.SalaryCurrency)
		case v.SalaryFrom != nil:
			salary = fmt.Sprintf("от %d %s", *v.SalaryFrom, v.SalaryCurrency)
		default:
			salary = fmt.Sprintf("до %d %s", *v.SalaryTo, v.SalaryCurrency)
		}
	}

	return fmt.Sprintf(
		"<b>%s</b>\n%s\nЗарплата: %s\nГород: %s\n%s",
		v.Name, v.EmployerName, salary, v.AreaName, v.URL,
	)
}
