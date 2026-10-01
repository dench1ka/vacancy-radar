package storage

import (
	"context"

	"github.com/dench1ka/vacancy-radar/internal/hhclient"
	"github.com/dench1ka/vacancy-radar/internal/worker"
)

// WorkerStore combines the subscription and vacancy stores behind the single
// interface the background worker depends on.
type WorkerStore struct {
	subscriptions *SubscriptionStore
	vacancies     *VacancyStore
}

func NewWorkerStore(subscriptions *SubscriptionStore, vacancies *VacancyStore) *WorkerStore {
	return &WorkerStore{subscriptions: subscriptions, vacancies: vacancies}
}

func (w *WorkerStore) ListActiveWithUsers(ctx context.Context) ([]worker.SubscriptionWithUser, error) {
	return w.subscriptions.ListActiveWithUsers(ctx)
}

func (w *WorkerStore) IsSeen(ctx context.Context, userID, hhVacancyID string) (bool, error) {
	return w.vacancies.IsSeen(ctx, userID, hhVacancyID)
}

func (w *WorkerStore) MarkSeen(ctx context.Context, userID string, v hhclient.Vacancy) error {
	return w.vacancies.MarkSeen(ctx, userID, v)
}
