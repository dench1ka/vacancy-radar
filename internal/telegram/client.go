// Package telegram sends notification messages via the Telegram Bot API.
package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func New(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.telegram.org",
	}
}

// SendMessage sends a text message to the given chat ID.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	if c.token == "" {
		return fmt.Errorf("telegram token is not configured")
	}

	reqURL := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.token)
	form := url.Values{}
	form.Set("chat_id", fmt.Sprintf("%d", chatID))
	form.Set("text", text)
	form.Set("parse_mode", "HTML")
	form.Set("disable_web_page_preview", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.URL.RawQuery = form.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}
	return nil
}
