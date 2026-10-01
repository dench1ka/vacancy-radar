package models

import "time"

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	TelegramChatID *int64    `json:"telegram_chat_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Subscription struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Keyword   string    `json:"keyword"`
	AreaID    string    `json:"area_id,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Vacancy struct {
	ID             string    `json:"id"`
	HHVacancyID    string    `json:"hh_vacancy_id"`
	Name           string    `json:"name"`
	EmployerName   string    `json:"employer_name"`
	URL            string    `json:"url"`
	SalaryFrom     *int      `json:"salary_from,omitempty"`
	SalaryTo       *int      `json:"salary_to,omitempty"`
	SalaryCurrency string    `json:"salary_currency,omitempty"`
	AreaName       string    `json:"area_name,omitempty"`
	PublishedAt    time.Time `json:"published_at"`
}
