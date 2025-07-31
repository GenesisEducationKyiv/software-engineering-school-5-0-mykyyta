package weatherhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"subscription/internal/domain"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

var ErrCityNotFound = domain.ErrCityNotFound

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type weatherResponse struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
}

func (c *Client) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	endpoint := fmt.Sprintf("%s/weather?city=%s", c.baseURL, url.QueryEscape(city))

	resp, err := c.doRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return domain.Report{}, fmt.Errorf("weather request failed: %w", err)
	}
	defer c.closeBody(ctx, resp.Body, "GetWeather")

	if resp.StatusCode == http.StatusNotFound {
		return domain.Report{}, ErrCityNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return domain.Report{}, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, resp.Status)
	}

	var res weatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return domain.Report{}, fmt.Errorf("decode weather response: %w", err)
	}

	return domain.Report{
		Temperature: res.Temperature,
		Humidity:    res.Humidity,
		Description: res.Description,
	}, nil
}

func (c *Client) CityIsValid(ctx context.Context, city string) (bool, error) {
	endpoint := fmt.Sprintf("%s/validate?city=%s", c.baseURL, url.QueryEscape(city))

	resp, err := c.doRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return false, fmt.Errorf("city validation request failed: %w", err)
	}
	defer c.closeBody(ctx, resp.Body, "CityIsValid")

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, resp.Status)
	}

	var data struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, fmt.Errorf("decode validation response: %w", err)
	}

	return data.Valid, nil
}

func (c *Client) doRequest(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if correlationID := loggerPkg.GetCorrelationID(ctx); correlationID != "" {
		req.Header.Set("X-Correlation-Id", correlationID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) closeBody(ctx context.Context, body io.Closer, operation string) {
	logger := loggerPkg.From(ctx)
	if err := body.Close(); err != nil {
		logger.Warn("Failed to close response body in %s: %v", operation, err)
	}
}
