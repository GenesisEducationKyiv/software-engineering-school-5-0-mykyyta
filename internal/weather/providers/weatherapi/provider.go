package weatherapi

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
	url := "https://api.weatherapi.com/v1"
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = baseURL[0]
	}
	return Provider{
		apiKey:  apiKey,
		baseURL: url,
	}
}

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
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (p Provider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	body, err := p.makeRequest(ctx, city)
	if err != nil {
		return weather.Report{}, err
	}

	if isCityNotFound(body) {
		return weather.Report{}, weather.ErrCityNotFound
	}

	var data weatherAPIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return weather.Report{}, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return weather.Report{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}, nil
}

func (p Provider) CityIsValid(ctx context.Context, city string) (bool, error) {
	body, err := p.makeRequest(ctx, city)
	if err != nil {
		return false, err
	}

	if isCityNotFound(body) {
		return false, weather.ErrCityNotFound
	}

	return true, nil
}

func (p Provider) makeRequest(ctx context.Context, city string) ([]byte, error) {
	if p.apiKey == "" {
		return nil, errors.New("missing API key")
	}
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", p.baseURL, p.apiKey, city)
	return doRequestBody(ctx, url)
}

func doRequestBody(ctx context.Context, url string) ([]byte, error) {
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
	return errResp.Error.Code == 1006
}

func closeBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Printf("failed to close response body: %v", err)
	}
}
