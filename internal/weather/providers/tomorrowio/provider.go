// internal/tomorrowio/provider.go
package tomorrowio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"weatherApi/internal/weather"
)

type Provider struct {
	apiKey  string
	baseURL string
}

func New(apiKey string, baseURL string) *Provider {
	if baseURL == "" {
		baseURL = "https://api.tomorrow.io/v4/weather/realtime"
	}
	return &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

func (p *Provider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	url := fmt.Sprintf("%s?location=%s&apikey=%s", p.baseURL, city, p.apiKey)
	body, err := makeRequest(ctx, url)
	if err != nil {
		return weather.Report{}, err
	}

	var res apiResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return weather.Report{}, fmt.Errorf("failed to decode tomorrow.io response: %w", err)
	}

	values := res.Data.Values
	return weather.Report{
		Temperature: values.Temperature,
		Humidity:    values.Humidity,
		Description: getDescription(values.WeatherCode),
	}, nil
}

func (p *Provider) CityIsValid(ctx context.Context, city string) (bool, error) {
	url := fmt.Sprintf("%s?location=%s&apikey=%s", p.baseURL, city, p.apiKey)
	body, err := makeRequest(ctx, url)
	if err != nil {
		return false, err
	}

	if isInvalidLocation(body) {
		return false, weather.ErrCityNotFound
	}
	return true, nil
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

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func isInvalidLocation(body []byte) bool {
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false
	}
	return errResp.Code == "InvalidLocation"
}

func makeRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return body, fmt.Errorf("API error: %s", body)
	}

	return io.ReadAll(resp.Body)
}

func getDescription(code int) string {
	switch code {
	case 1000:
		return "Clear"
	case 1101:
		return "Partly Cloudy"
	case 1102:
		return "Mostly Cloudy"
	case 4000:
		return "Drizzle"
	case 4200:
		return "Light Rain"
	case 5000:
		return "Snow"
	default:
		return fmt.Sprintf("Code %d", code)
	}
}
