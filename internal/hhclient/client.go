// Package hhclient provides a minimal client for the public hh.ru vacancy search API.
package hhclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type SearchParams struct {
	Text    string
	AreaID  string
	PerPage int
}

type searchResponse struct {
	Items []rawVacancy `json:"items"`
}

type rawVacancy struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AlternateURL string `json:"alternate_url"`
	PublishedAt  string `json:"published_at"`
	Employer     struct {
		Name string `json:"name"`
	} `json:"employer"`
	Area struct {
		Name string `json:"name"`
	} `json:"area"`
	Salary *struct {
		From     *int   `json:"from"`
		To       *int   `json:"to"`
		Currency string `json:"currency"`
	} `json:"salary"`
}

type Vacancy struct {
	HHVacancyID    string
	Name           string
	EmployerName   string
	URL            string
	AreaName       string
	SalaryFrom     *int
	SalaryTo       *int
	SalaryCurrency string
	PublishedAt    time.Time
}

// Search queries the hh.ru vacancy search endpoint and returns a normalized list of vacancies.
func (c *Client) Search(ctx context.Context, p SearchParams) ([]Vacancy, error) {
	if p.PerPage <= 0 {
		p.PerPage = 20
	}

	q := url.Values{}
	q.Set("text", p.Text)
	q.Set("per_page", fmt.Sprintf("%d", p.PerPage))
	q.Set("order_by", "publication_time")
	if p.AreaID != "" {
		q.Set("area", p.AreaID)
	}

	reqURL := fmt.Sprintf("%s/vacancies?%s", c.baseURL, q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "vacancy-radar/1.0 (+https://github.com/dench1ka/vacancy-radar)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hh.ru returned status %d", resp.StatusCode)
	}

	var parsed searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	vacancies := make([]Vacancy, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		v := Vacancy{
			HHVacancyID:  item.ID,
			Name:         item.Name,
			EmployerName: item.Employer.Name,
			URL:          item.AlternateURL,
			AreaName:     item.Area.Name,
		}
		if item.Salary != nil {
			v.SalaryFrom = item.Salary.From
			v.SalaryTo = item.Salary.To
			v.SalaryCurrency = item.Salary.Currency
		}
		if t, err := parsePublishedAt(item.PublishedAt); err == nil {
			v.PublishedAt = t
		}
		vacancies = append(vacancies, v)
	}

	return vacancies, nil
}

// hh.ru returns timestamps like "2014-07-22T15:38:17+0400" — a numeric offset
// without a colon, which time.RFC3339 does not accept.
const hhTimeLayout = "2006-01-02T15:04:05-0700"

func parsePublishedAt(s string) (time.Time, error) {
	if t, err := time.Parse(hhTimeLayout, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
