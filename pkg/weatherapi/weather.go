package weatherapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"weatherApi/internal/model"
)

// weatherAPIResponse defines the structure of the external API response (weatherapi.com).
// Used internally to decode the raw JSON before mapping to our model.
type weatherAPIResponse struct {
	Current struct {
		TempC     float64 `json:"temp_c"`
		Humidity  int     `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

// FetchWithStatus retrieves current weather for the given city from weatherapi.com.
// Returns a pointer to Weather model, HTTP-like status code, and error if any.
// This function is used in both API responses and email updates.
func FetchWithStatus(city string) (*model.Weather, int, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return nil, http.StatusInternalServerError, fmt.Errorf("weather API key not set")
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)
	resp, err := http.Get(url)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("failed to close response body: %v", cerr)
		}
	}()

	switch resp.StatusCode {
	case 400:
		return nil, http.StatusBadRequest, fmt.Errorf("Invalid city name")
	case 404:
		return nil, http.StatusNotFound, fmt.Errorf("City not found")
	case 200:
		// OK — continue parsing
	default:
		return nil, http.StatusBadGateway, fmt.Errorf("Weather API returned unexpected status")
	}

	var data weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Failed to parse weather data")
	}

	// Map response data to internal model
	result := &model.Weather{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}

	return result, http.StatusOK, nil
}

// CityExists checks whether the specified city is valid using the external API.
// Used during subscription to validate user input before storing in DB.
// Returns false for 400/404, true for 200, and error for any other status.
func CityExists(city string) (bool, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return false, fmt.Errorf("weather API key not set")
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("weather API request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected weather API response: %s", resp.Status)
}
