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

	"weather/internal/domain"
)

type Provider struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

func New(apiKey string, client *http.Client, baseURL ...string) Provider {
	url := "https://api.weatherapi.com/v1"
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

type weatherAPIResponse struct {
	Current struct {
		TempC     float64 `json:"temp_c"` //nolint:tagliatelle
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

func (p Provider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	body, err := p.makeRequest(ctx, city)
	if err != nil {
		return domain.Report{}, err
	}

	if isCityNotFound(body) {
		return domain.Report{}, domain.ErrCityNotFound
	}

	var data weatherAPIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return domain.Report{}, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return domain.Report{
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
		return false, domain.ErrCityNotFound
	}

	return true, nil
}

func (p Provider) makeRequest(ctx context.Context, city string) ([]byte, error) {
	if p.apiKey == "" {
		return nil, errors.New("missing API key")
	}
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", p.baseURL, p.apiKey, city)
	return p.doRequestBody(ctx, url)
}

func (p Provider) doRequestBody(ctx context.Context, url string) ([]byte, error) {
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
