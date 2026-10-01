package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dench1ka/vacancy-radar/internal/hhclient"
	"github.com/dench1ka/vacancy-radar/internal/models"
	"github.com/google/uuid"
)

type VacancyStore struct {
	db *sql.DB
}

func NewVacancyStore(db *sql.DB) *VacancyStore {
	return &VacancyStore{db: db}
}

// IsSeen reports whether this vacancy has already been recorded as sent to the given user.
func (s *VacancyStore) IsSeen(ctx context.Context, userID, hhVacancyID string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM seen_vacancies WHERE user_id = $1 AND hh_vacancy_id = $2)`
	var exists bool
	if err := s.db.QueryRowContext(ctx, q, userID, hhVacancyID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check seen vacancy: %w", err)
	}
	return exists, nil
}

// MarkSeen upserts the vacancy details into the shared cache and records that it was
// sent to the given user, so future polls won't notify them about it again.
func (s *VacancyStore) MarkSeen(ctx context.Context, userID string, v hhclient.Vacancy) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const upsertVacancy = `
		INSERT INTO vacancies (id, hh_vacancy_id, name, employer_name, url, area_name, salary_from, salary_to, salary_currency, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (hh_vacancy_id) DO UPDATE SET
			name = EXCLUDED.name,
			employer_name = EXCLUDED.employer_name,
			url = EXCLUDED.url`

	_, err = tx.ExecContext(ctx, upsertVacancy,
		uuid.NewString(), v.HHVacancyID, v.Name, v.EmployerName, v.URL, v.AreaName,
		v.SalaryFrom, v.SalaryTo, v.SalaryCurrency, v.PublishedAt)
	if err != nil {
		return fmt.Errorf("upsert vacancy: %w", err)
	}

	const insertSeen = `
		INSERT INTO seen_vacancies (user_id, hh_vacancy_id, sent_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, hh_vacancy_id) DO NOTHING`

	if _, err := tx.ExecContext(ctx, insertSeen, userID, v.HHVacancyID, time.Now()); err != nil {
		return fmt.Errorf("insert seen vacancy: %w", err)
	}

	return tx.Commit()
}

// ListForUser returns vacancies previously sent to the user, most recent first.
func (s *VacancyStore) ListForUser(ctx context.Context, userID string, limit, offset int) ([]models.Vacancy, error) {
	const q = `
		SELECT v.id, v.hh_vacancy_id, v.name, v.employer_name, v.url, v.area_name,
		       v.salary_from, v.salary_to, v.salary_currency, v.published_at
		FROM seen_vacancies sv
		JOIN vacancies v ON v.hh_vacancy_id = sv.hh_vacancy_id
		WHERE sv.user_id = $1
		ORDER BY sv.sent_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list vacancies: %w", err)
	}
	defer rows.Close()

	result := []models.Vacancy{}
	for rows.Next() {
		var v models.Vacancy
		if err := rows.Scan(&v.ID, &v.HHVacancyID, &v.Name, &v.EmployerName, &v.URL, &v.AreaName,
			&v.SalaryFrom, &v.SalaryTo, &v.SalaryCurrency, &v.PublishedAt); err != nil {
			return nil, fmt.Errorf("scan vacancy: %w", err)
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
