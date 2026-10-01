package hhclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const sampleResponse = `{
  "items": [
    {
      "id": "12345",
      "name": "Junior Golang Developer",
      "alternate_url": "https://hh.ru/vacancy/12345",
      "published_at": "2026-09-30T10:00:00+0300",
      "employer": {"name": "Acme Corp"},
      "area": {"name": "Moscow"},
      "salary": {"from": 80000, "to": 120000, "currency": "RUR"}
    },
    {
      "id": "67890",
      "name": "Go Developer (no salary)",
      "alternate_url": "https://hh.ru/vacancy/67890",
      "published_at": "2026-09-29T10:00:00+0300",
      "employer": {"name": "Beta LLC"},
      "area": {"name": "Remote"},
      "salary": null
    }
  ]
}`

func TestClient_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("text") != "golang junior" {
			t.Errorf("expected text=golang junior, got %s", r.URL.Query().Get("text"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleResponse))
	}))
	defer server.Close()

	client := New(server.URL)
	vacancies, err := client.Search(context.Background(), SearchParams{Text: "golang junior"})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(vacancies) != 2 {
		t.Fatalf("expected 2 vacancies, got %d", len(vacancies))
	}

	first := vacancies[0]
	if first.HHVacancyID != "12345" || first.Name != "Junior Golang Developer" {
		t.Errorf("unexpected first vacancy: %+v", first)
	}
	if first.SalaryFrom == nil || *first.SalaryFrom != 80000 {
		t.Errorf("expected salary from 80000, got %+v", first.SalaryFrom)
	}
	if first.PublishedAt.IsZero() {
		t.Error("expected PublishedAt to be parsed from hh.ru's numeric-offset timestamp, got zero value")
	}

	second := vacancies[1]
	if second.SalaryFrom != nil {
		t.Errorf("expected nil salary for second vacancy, got %+v", second.SalaryFrom)
	}
}

func TestClient_Search_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(server.URL)
	if _, err := client.Search(context.Background(), SearchParams{Text: "golang"}); err == nil {
		t.Error("expected error for non-200 response")
	}
}
