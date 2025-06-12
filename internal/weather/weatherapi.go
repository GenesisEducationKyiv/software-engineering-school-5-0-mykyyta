package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type WeatherApiProvider struct {
	apiKey string
}

func NewWeatherAPIProvider(apiKey string) Provider {
	return &WeatherApiProvider{apiKey: apiKey}
}

type weatherAPIResponse struct {
	Current struct {
		TempC     float64 `json:"tempC"`
		Humidity  int     `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

func (p *WeatherApiProvider) GetCurrentWeather(ctx context.Context, city string) (*Weather, error) {
	if p.apiKey == "" {
		return nil, errors.New("weather API key not set")
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", p.apiKey, city)
	resp, err := doRequestWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API error: %s", resp.Status)
	}

	var data weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return &Weather{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}, nil
}

func (p *WeatherApiProvider) CityExists(ctx context.Context, city string) (bool, error) {
	if p.apiKey == "" {
		return false, errors.New("weather API key not set")
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", p.apiKey, city)
	resp, err := doRequestWithContext(ctx, url)
	if err != nil {
		return false, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusBadRequest, http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected weather API response: %s", resp.Status)
	}
}

func doRequestWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timed out: %w", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func closeBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Printf("failed to close response body: %v", err)
	}
}
