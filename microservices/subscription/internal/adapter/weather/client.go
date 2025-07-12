package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"subscription/internal/domain"
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

func (c *Client) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	endpoint := fmt.Sprintf("%s/weather?city=%s", c.baseURL, url.QueryEscape(city))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.Report{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.Report{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Report{}, ErrCityNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return domain.Report{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var data struct {
		City        string  `json:"city"`
		Temperature float64 `json:"temperature"`
		Humidity    int     `json:"humidity"`
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return domain.Report{}, fmt.Errorf("decode response: %w", err)
	}

	return domain.Report{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Description: data.Description,
	}, nil
}

func (c *Client) CityIsValid(ctx context.Context, city string) (bool, error) {
	endpoint := fmt.Sprintf("%s/validate?city=%s", c.baseURL, url.QueryEscape(city))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var data struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}

	return data.Valid, nil
}
