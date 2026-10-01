package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dench1ka/vacancy-radar/internal/hhclient"
)

type fakeStore struct {
	mu   sync.Mutex
	subs []SubscriptionWithUser
	seen map[string]bool // key: userID+"|"+hhVacancyID
	sent []hhclient.Vacancy
}

func newFakeStore(subs []SubscriptionWithUser) *fakeStore {
	return &fakeStore{subs: subs, seen: map[string]bool{}}
}

func (f *fakeStore) ListActiveWithUsers(ctx context.Context) ([]SubscriptionWithUser, error) {
	return f.subs, nil
}

func (f *fakeStore) IsSeen(ctx context.Context, userID, hhVacancyID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.seen[userID+"|"+hhVacancyID], nil
}

func (f *fakeStore) MarkSeen(ctx context.Context, userID string, v hhclient.Vacancy) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen[userID+"|"+v.HHVacancyID] = true
	f.sent = append(f.sent, v)
	return nil
}

type fakeHH struct {
	vacancies []hhclient.Vacancy
}

func (f *fakeHH) Search(ctx context.Context, p hhclient.SearchParams) ([]hhclient.Vacancy, error) {
	return f.vacancies, nil
}

type fakeNotifier struct {
	mu       sync.Mutex
	messages []string
}

func (f *fakeNotifier) SendMessage(ctx context.Context, chatID int64, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = append(f.messages, text)
	return nil
}

func TestWorker_RunOnce_NotifiesNewVacancyOnce(t *testing.T) {
	chatID := int64(42)
	sub := SubscriptionWithUser{SubscriptionID: "s1", UserID: "u1", Keyword: "golang junior", TelegramChatID: &chatID}
	store := newFakeStore([]SubscriptionWithUser{sub})
	hh := &fakeHH{vacancies: []hhclient.Vacancy{{HHVacancyID: "v1", Name: "Junior Go Dev"}}}
	notifier := &fakeNotifier{}

	w := New(store, hh, notifier, time.Minute, nil)

	if err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if len(notifier.messages) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.messages))
	}

	// Running again should not re-notify about the same vacancy.
	if err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("second RunOnce returned error: %v", err)
	}
	if len(notifier.messages) != 1 {
		t.Fatalf("expected still 1 notification after second run, got %d", len(notifier.messages))
	}
}

func TestWorker_RunOnce_SkipsNotificationWithoutTelegramChatID(t *testing.T) {
	sub := SubscriptionWithUser{SubscriptionID: "s1", UserID: "u1", Keyword: "golang junior", TelegramChatID: nil}
	store := newFakeStore([]SubscriptionWithUser{sub})
	hh := &fakeHH{vacancies: []hhclient.Vacancy{{HHVacancyID: "v1", Name: "Junior Go Dev"}}}
	notifier := &fakeNotifier{}

	w := New(store, hh, notifier, time.Minute, nil)

	if err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if len(notifier.messages) != 0 {
		t.Fatalf("expected no notifications, got %d", len(notifier.messages))
	}
	if len(store.sent) != 1 {
		t.Fatalf("expected vacancy to still be marked seen, got %d", len(store.sent))
	}
}

func TestWorker_RunOnce_ProcessesSubscriptionsConcurrently(t *testing.T) {
	var subs []SubscriptionWithUser
	for i := 0; i < 20; i++ {
		chatID := int64(i)
		subs = append(subs, SubscriptionWithUser{
			SubscriptionID: "s", UserID: "u" + string(rune('a'+i)), Keyword: "golang", TelegramChatID: &chatID,
		})
	}
	store := newFakeStore(subs)
	hh := &fakeHH{vacancies: []hhclient.Vacancy{{HHVacancyID: "v1"}}}
	notifier := &fakeNotifier{}

	w := New(store, hh, notifier, time.Minute, nil)

	if err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if len(notifier.messages) != 20 {
		t.Fatalf("expected 20 notifications, got %d", len(notifier.messages))
	}
}
