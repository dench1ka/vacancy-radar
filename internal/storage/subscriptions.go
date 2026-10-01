package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dench1ka/vacancy-radar/internal/models"
	"github.com/dench1ka/vacancy-radar/internal/worker"
	"github.com/google/uuid"
)

type SubscriptionStore struct {
	db *sql.DB
}

func NewSubscriptionStore(db *sql.DB) *SubscriptionStore {
	return &SubscriptionStore{db: db}
}

func (s *SubscriptionStore) Create(ctx context.Context, userID, keyword, areaID string) (models.Subscription, error) {
	id := uuid.NewString()
	const q = `
		INSERT INTO subscriptions (id, user_id, keyword, area_id, is_active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING created_at`

	sub := models.Subscription{ID: id, UserID: userID, Keyword: keyword, AreaID: areaID, IsActive: true}
	err := s.db.QueryRowContext(ctx, q, id, userID, keyword, areaID).Scan(&sub.CreatedAt)
	if err != nil {
		return models.Subscription{}, fmt.Errorf("insert subscription: %w", err)
	}
	return sub, nil
}

func (s *SubscriptionStore) ListByUser(ctx context.Context, userID string) ([]models.Subscription, error) {
	const q = `SELECT id, user_id, keyword, area_id, is_active, created_at FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	subs := []models.Subscription{}
	for rows.Next() {
		var sub models.Subscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Keyword, &sub.AreaID, &sub.IsActive, &sub.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

func (s *SubscriptionStore) Delete(ctx context.Context, userID, id string) error {
	const q = `DELETE FROM subscriptions WHERE id = $1 AND user_id = $2`
	res, err := s.db.ExecContext(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListActiveWithUsers returns every active subscription joined with the owning user's
// Telegram chat ID, for the background worker to poll against hh.ru.
func (s *SubscriptionStore) ListActiveWithUsers(ctx context.Context) ([]worker.SubscriptionWithUser, error) {
	const q = `
		SELECT s.id, s.user_id, s.keyword, s.area_id, u.telegram_chat_id
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		WHERE s.is_active = true`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list active subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []worker.SubscriptionWithUser
	for rows.Next() {
		var sub worker.SubscriptionWithUser
		if err := rows.Scan(&sub.SubscriptionID, &sub.UserID, &sub.Keyword, &sub.AreaID, &sub.TelegramChatID); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}
