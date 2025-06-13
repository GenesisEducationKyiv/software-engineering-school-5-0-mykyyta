package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type WeatherApiProvider struct {
	apiKey string
}

func NewWeatherAPIProvider(apiKey string) Provider {
	return &WeatherApiProvider{apiKey: apiKey}
}

var ErrCityNotFound = errors.New("city not found")

type weatherAPIResponse struct {
	Current struct {
		TempC     float64 `json:"temp_с"` //nolint:tagliatelle
		Humidity  int     `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

type errorAPIResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *WeatherApiProvider) GetCurrentWeather(ctx context.Context, city string) (*Weather, error) {
	body, err := makeWeatherAPIRequest(ctx, p.apiKey, city)
	if err != nil {
		return nil, err
	}

	if isCityNotFound(body) {
		return nil, ErrCityNotFound
	}

	var data weatherAPIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return &Weather{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}, nil
}

func (p *WeatherApiProvider) CityExists(ctx context.Context, city string) (bool, error) {
	body, err := makeWeatherAPIRequest(ctx, p.apiKey, city)
	if err != nil {
		return false, err
	}

	if isCityNotFound(body) {
		return false, nil
	}

	return true, nil
}

func makeWeatherAPIRequest(ctx context.Context, apiKey, city string) ([]byte, error) {
	if apiKey == "" {
		return nil, errors.New("missing API key")
	}
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)
	return doRequestBody(ctx, url)
}

func doRequestBody(ctx context.Context, url string) ([]byte, error) {
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
	defer closeBody(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func isCityNotFound(body []byte) bool {
	var errResp errorAPIResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false
	}
	return errResp.Error.Message == "No matching location found."
}

func closeBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Printf("failed to close response body: %v", err)
	}
}
