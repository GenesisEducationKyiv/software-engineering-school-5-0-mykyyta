package openweathermap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"weatherApi/internal/weather"
)

type Provider struct {
	apiKey  string
	baseURL string
}

func New(apiKey string, baseURL ...string) Provider {
	url := "https://api.openweathermap.org/data/2.5/weather"
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = baseURL[0]
	}
	return Provider{
		apiKey:  apiKey,
		baseURL: url,
	}
}

type apiResponse struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Cod int `json:"cod"`
}

func (p Provider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	url := fmt.Sprintf("%s?q=%s&appid=%s&units=metric", p.baseURL, city, p.apiKey)
	body, err := makeRequest(ctx, url)
	if err != nil {
		if isCityNotFound(body) {
			return weather.Report{}, weather.ErrCityNotFound
		}
		return weather.Report{}, err
	}

	var res apiResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return weather.Report{}, fmt.Errorf("failed to decode OpenWeatherMap response: %w", err)
	}

	if len(res.Weather) == 0 {
		return weather.Report{}, errors.New("missing weather description")
	}

	return weather.Report{
		Temperature: res.Main.Temp,
		Humidity:    res.Main.Humidity,
		Description: res.Weather[0].Description,
	}, nil
}

func (p Provider) CityIsValid(ctx context.Context, city string) (bool, error) {
	url := fmt.Sprintf("%s?q=%s&appid=%s", p.baseURL, city, p.apiKey)
	body, err := makeRequest(ctx, url)
	if err != nil {
		if isCityNotFound(body) {
			return false, weather.ErrCityNotFound
		}
		return false, err
	}
	return true, nil
}

func isCityNotFound(body []byte) bool {
	var errResp struct {
		Cod     string `json:"cod"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false
	}
	return errResp.Cod == "404"
}

func makeRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timed out: %w", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer closeBody(resp)

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("API error: %s", body)
	}

	return body, nil
}

func closeBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Printf("failed to close response body: %v", err)
	}
}
