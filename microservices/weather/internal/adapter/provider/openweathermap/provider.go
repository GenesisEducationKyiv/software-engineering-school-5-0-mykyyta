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

	"weather/internal/domain"
)

type Provider struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

func New(apiKey string, client *http.Client, baseURL ...string) Provider {
	url := "https://api.openweathermap.org/data/2.5/weather"
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
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Cod int `json:"cod"`
}

func (p Provider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	url := fmt.Sprintf("%s?q=%s&appid=%s&units=metric", p.baseURL, city, p.apiKey)
	body, err := p.makeRequest(ctx, url)
	if err != nil {
		if isCityNotFound(body) {
			return domain.Report{}, domain.ErrCityNotFound
		}
		return domain.Report{}, err
	}

	var res apiResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return domain.Report{}, fmt.Errorf("failed to decode OpenWeatherMap response: %w", err)
	}

	if len(res.Weather) == 0 {
		return domain.Report{}, errors.New("missing weather description")
	}

	return domain.Report{
		Temperature: res.Main.Temp,
		Humidity:    res.Main.Humidity,
		Description: res.Weather[0].Description,
	}, nil
}

func (p Provider) CityIsValid(ctx context.Context, city string) (bool, error) {
	url := fmt.Sprintf("%s?q=%s&appid=%s", p.baseURL, city, p.apiKey)
	body, err := p.makeRequest(ctx, url)
	if err != nil {
		if isCityNotFound(body) {
			return false, domain.ErrCityNotFound
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
