package weatherlogger

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"weatherApi/internal/weather"
)

type fakeProvider struct{}

func (f *fakeProvider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	return weather.Report{
		Temperature: 20.5,
		Humidity:    60,
		Description: "Sunny",
	}, nil
}

func (f *fakeProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return true, nil
}

func TestLogger_GetWeather_LogsOutput(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf) // Перехоплюємо log.Print

	ctx := context.Background()
	fake := &fakeProvider{}
	logger := New(fake, "fake")

	_, err := logger.GetWeather(ctx, "Kyiv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logContent := buf.String()
	if !strings.Contains(logContent, "GetWeather") || !strings.Contains(logContent, "Kyiv") {
		t.Errorf("expected log to contain method and city, got: %q", logContent)
	}
}

func TestLogger_CityIsValid_LogsOutput(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	ctx := context.Background()
	fake := &fakeProvider{}
	logger := New(fake, "fake")

	_, err := logger.CityIsValid(ctx, "London")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logContent := buf.String()
	if !strings.Contains(logContent, "CityIsValid") || !strings.Contains(logContent, "London") {
		t.Errorf("expected log to contain method and city, got: %q", logContent)
	}
}
