package tomorrowio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"weather/internal/domain"
)

type Provider struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

func New(apiKey string, client *http.Client, baseURL ...string) Provider {
	url := "https://api.tomorrow.io/v4/weather/realtime"
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = baseURL[0]
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return Provider{
		apiKey:  apiKey,
		client:  client,
		baseURL: url,
	}
}

type apiResponse struct {
	Data struct {
		Values struct {
			Temperature float64 `json:"temperature"`
			Humidity    int     `json:"humidity"`
			WeatherCode int     `json:"weatherCode"`
		} `json:"values"`
	} `json:"data"`
}

func (p Provider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	url := fmt.Sprintf("%s?location=%s&apikey=%s", p.baseURL, city, p.apiKey)
	body, err := p.makeRequest(ctx, url)
	if err != nil {
		if isInvalidLocation(body) {
			return domain.Report{}, domain.ErrCityNotFound
		}
		return domain.Report{}, err
	}

	var res apiResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return domain.Report{}, fmt.Errorf("failed to decode tomorrow.io response: %w", err)
	}

	values := res.Data.Values
	return domain.Report{
		Temperature: values.Temperature,
		Humidity:    values.Humidity,
		Description: getDescription(values.WeatherCode),
	}, nil
}

func (p Provider) CityIsValid(ctx context.Context, city string) (bool, error) {
	url := fmt.Sprintf("%s?location=%s&apikey=%s", p.baseURL, city, p.apiKey)
	body, err := p.makeRequest(ctx, url)
	if err != nil {
		if isInvalidLocation(body) {
			return false, domain.ErrCityNotFound
		}
		return false, err
	}
	return true, nil
}

func isInvalidLocation(body []byte) bool {
	var errResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false
	}
	return errResp.Code == 400001
}

func (p Provider) makeRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timed out: %w", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return body, fmt.Errorf("API error: %s", body)
	}

	return io.ReadAll(resp.Body)
}

func getDescription(code int) string {
	if desc, ok := weatherCodeDescriptions[code]; ok {
		return desc
	}
	return fmt.Sprintf("Unknown (code %d)", code)
}

func closeBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Printf("failed to close response body: %v", err)
	}
}
