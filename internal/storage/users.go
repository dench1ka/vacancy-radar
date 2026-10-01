package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dench1ka/vacancy-radar/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, email, passwordHash string) (models.User, error) {
	id := uuid.NewString()
	const q = `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3) RETURNING created_at`

	var u models.User
	u.ID = id
	u.Email = email
	u.PasswordHash = passwordHash

	err := s.db.QueryRowContext(ctx, q, id, email, passwordHash).Scan(&u.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrAlreadyExists
		}
		return models.User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (models.User, error) {
	const q = `SELECT id, email, password_hash, telegram_chat_id, created_at FROM users WHERE email = $1`
	var u models.User
	err := s.db.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TelegramChatID, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (s *UserStore) GetByID(ctx context.Context, id string) (models.User, error) {
	const q = `SELECT id, email, password_hash, telegram_chat_id, created_at FROM users WHERE id = $1`
	var u models.User
	err := s.db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TelegramChatID, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (s *UserStore) SetTelegramChatID(ctx context.Context, userID string, chatID int64) error {
	const q = `UPDATE users SET telegram_chat_id = $1 WHERE id = $2`
	res, err := s.db.ExecContext(ctx, q, chatID, userID)
	if err != nil {
		return fmt.Errorf("update telegram chat id: %w", err)
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

const pgUniqueViolationCode = "23505"

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode
}
