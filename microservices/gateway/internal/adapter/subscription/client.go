package subscription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SubscriptionClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSubscriptionClient(baseURL string, timeout time.Duration) *SubscriptionClient {
	return &SubscriptionClient{
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
}

type ConfirmRequest struct {
	Token string `json:"token"`
}

type UnsubscribeRequest struct {
	Token string `json:"token"`
}

func (c *SubscriptionClient) Subscribe(ctx context.Context, req SubscribeRequest) (*SubscribeResponse, error) {
	return c.post(ctx, "/api/subscription", req, &SubscribeResponse{})
}

func (c *SubscriptionClient) Confirm(ctx context.Context, req ConfirmRequest) (*SubscribeResponse, error) {
	return c.post(ctx, "/api/subscription/confirm", req, &SubscribeResponse{})
}

func (c *SubscriptionClient) Unsubscribe(ctx context.Context, req UnsubscribeRequest) (*SubscribeResponse, error) {
	return c.post(ctx, "/api/subscription/unsubscribe", req, &SubscribeResponse{})
}

func (c *SubscriptionClient) post(ctx context.Context, endpoint string, reqBody interface{}, respBody interface{}) (*SubscribeResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "api-gateway/1.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer c.closeBody(resp.Body)

	// Детальна обробка статусів
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return respBody.(*SubscribeResponse), nil
}

func (c *SubscriptionClient) closeBody(body io.Closer) {
	if err := body.Close(); err != nil {
		fmt.Printf("Failed to close response body: %v\n", err)
	}
}
