package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"subscription/internal/domain"
)

type Client struct {
	baseURL string
	logger  *log.Logger
	client  *http.Client
}

func NewClient(baseURL string, logger *log.Logger, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	return &Client{
		baseURL: baseURL,
		logger:  logger,
		client:  client,
	}
}

func (e *Client) SendConfirmationEmail(ctx context.Context, email, token string) error {
	return e.send(ctx, Request{
		To:       email,
		Template: "confirmation",
		Data: map[string]string{
			"token": token,
		},
	})
}

func (e *Client) SendWeatherReport(ctx context.Context, email string, weather domain.Report, city, token string) error {
	return e.send(ctx, Request{
		To:       email,
		Template: "weather_report",
		Data: map[string]string{
			"temperature": fmt.Sprintf("%.2f", weather.Temperature),
			"humidity":    fmt.Sprintf("%d", weather.Humidity),
			"description": weather.Description,
			"city":        city,
			"token":       token,
		},
	})
}

func (e *Client) send(ctx context.Context, req Request) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	url := fmt.Sprintf("%s/send", e.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("email service returned status %d", resp.StatusCode)
	}

	e.logger.Printf("Email sent to %s with template %s", req.To, req.Template)
	return nil
}

type Request struct {
	To       string            `json:"to"`
	Template string            `json:"template"`
	Data     map[string]string `json:"data"`
}
