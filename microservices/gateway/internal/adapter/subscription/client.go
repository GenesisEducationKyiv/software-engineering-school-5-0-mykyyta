package subscription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gateway/internal/middleware"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type SubscribeRequest struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
}

type SubscribeResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type ConfirmResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type UnsubscribeResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type WeatherResponse struct {
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
	Humidity    int     `json:"humidity"`
}

func (c *Client) Subscribe(ctx context.Context, req SubscribeRequest) (*SubscribeResponse, error) {
	var resp SubscribeResponse
	err := c.postJSON(ctx, "/api/subscribe", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Confirm(ctx context.Context, token string) (*ConfirmResponse, error) {
	endpoint := fmt.Sprintf("/api/confirm/%s", url.PathEscape(token))
	var resp ConfirmResponse
	err := c.get(ctx, endpoint, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Unsubscribe(ctx context.Context, token string) (*UnsubscribeResponse, error) {
	endpoint := fmt.Sprintf("/api/unsubscribe/%s", url.PathEscape(token))
	var resp UnsubscribeResponse
	err := c.get(ctx, endpoint, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetWeather(ctx context.Context, city string) (*WeatherResponse, error) {
	endpoint := "/api/weather"
	if city != "" {
		endpoint += "?city=" + url.QueryEscape(city)
	}
	var resp WeatherResponse
	err := c.get(ctx, endpoint, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) postJSON(ctx context.Context, endpoint string, reqBody interface{}, respBody interface{}) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	fullURL := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "api-gateway/1.0.0")

	if requestID := middleware.GetRequestID(ctx); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer c.closeBody(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) get(ctx context.Context, endpoint string, respBody interface{}) error {
	fullURL := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "api-gateway/1.0.0")

	if requestID := middleware.GetRequestID(ctx); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer c.closeBody(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) closeBody(body io.Closer) {
	if err := body.Close(); err != nil {
		fmt.Printf("Failed to close response body: %v\n", err)
	}
}
