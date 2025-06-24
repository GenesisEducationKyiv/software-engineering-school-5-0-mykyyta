//go:build integration

package weather_test

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"weatherApi/internal/weather"
)

// --- Fake Providers ---

type successProvider struct{}

func (s *successProvider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	return weather.Report{
		Temperature: 20.5,
		Humidity:    60,
		Description: "Sunny",
	}, nil
}

func (s *successProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return true, nil
}

type failProvider struct{}

func (f *failProvider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	return weather.Report{}, context.DeadlineExceeded
}

func (f *failProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return false, context.DeadlineExceeded
}

// --- Test ---

func TestIntegration_ChainAndLogging(t *testing.T) {
	logFile := "test_weather.log"
	defer func() {
		if err := os.Remove(logFile); err != nil {
			t.Logf("failed to remove log file: %v", err)
		}
	}()

	file, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("could not create log file: %v", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			t.Logf("failed to close file: %v", err)
		}
	}()

	logger := log.New(file, "", log.LstdFlags)

	fp := &failProvider{}
	sp := &successProvider{}
	wrappedFail := weather.NewWrapper(fp, "FailProvider", logger)
	wrappedSuccess := weather.NewWrapper(sp, "SuccessProvider", logger)

	chain := weather.NewChain(wrappedFail, wrappedSuccess)

	ctx := context.Background()
	report, err := chain.GetWeather(ctx, "Kyiv")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if report.Description != "Sunny" {
		t.Errorf("expected 'Sunny', got: %s", report.Description)
	}

	if err := file.Sync(); err != nil {
		t.Fatalf("failed to sync file: %v", err)
	}
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "GetWeather") || !strings.Contains(content, "Kyiv") {
		t.Errorf("log does not contain expected entries, got:\n%s", content)
	}
}

func TestIntegration_ChainErrors(t *testing.T) {
	ctx := context.Background()

	fp1 := &failProvider{}
	fp2 := &failProvider{} // другий, і його помилка буде останньою
	chain := weather.NewChain(fp1, fp2)

	_, err := chain.GetWeather(ctx, "Nowhere")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedPrefix := "all Providers failed: "
	if !strings.HasPrefix(err.Error(), expectedPrefix) {
		t.Errorf("expected error to start with %q, got: %v", expectedPrefix, err)
	}

	expectedSuffix := context.DeadlineExceeded.Error()
	if !strings.HasSuffix(err.Error(), expectedSuffix) {
		t.Errorf("expected last error to be %q, got: %v", expectedSuffix, err)
	}
}

type cityNotFoundProvider struct{}

func (c *cityNotFoundProvider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	return weather.Report{}, weather.ErrCityNotFound
}

func (c *cityNotFoundProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return false, weather.ErrCityNotFound
}

func TestIntegration_CityNotFoundError(t *testing.T) {
	ctx := context.Background()

	providers := []weather.Provider{
		&failProvider{},
		&cityNotFoundProvider{},
		&failProvider{},
	}
	chain := weather.NewChain(providers...)

	ok, err := chain.CityIsValid(ctx, "Atlantis")
	if ok {
		t.Errorf("expected city to be invalid, got ok=true")
	}
	if !errors.Is(err, weather.ErrCityNotFound) {
		t.Errorf("expected ErrCityNotFound, got: %v", err)
	}
}
