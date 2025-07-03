package weather

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"weatherApi/internal/domain"
)

type fakeProvider struct{}

func (f *fakeProvider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	return domain.Report{
		Temperature: 25.0,
		Humidity:    55,
		Description: "Cloudy",
	}, nil
}

func (f *fakeProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return true, nil
}

type FakeLogger struct {
	buf bytes.Buffer
}

func (f *FakeLogger) Write(p []byte) (n int, err error) {
	return f.buf.Write(p)
}

func (f *FakeLogger) String() string {
	return f.buf.String()
}

func TestLogger_GetWeather_LogsOutput(t *testing.T) {
	fakeLog := &FakeLogger{}
	logger := log.New(fakeLog, "", 0)

	fake := &fakeProvider{}
	logged := NewLogWrapper(fake, "FakeWeather", logger)

	ctx := context.Background()
	_, err := logged.GetWeather(ctx, "Kyiv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logContent := fakeLog.String()
	if !strings.Contains(logContent, "GetWeather") || !strings.Contains(logContent, "Kyiv") {
		t.Errorf("expected log to contain method and city, got: %q", logContent)
	}
}

func TestLogger_CityIsValid_LogsOutput(t *testing.T) {
	fakeLog := &FakeLogger{}
	logger := log.New(fakeLog, "", 0)

	fake := &fakeProvider{}
	logged := NewLogWrapper(fake, "FakeWeather", logger)

	ctx := context.Background()
	_, err := logged.CityIsValid(ctx, "London")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logContent := fakeLog.String()
	if !strings.Contains(logContent, "CityIsValid") || !strings.Contains(logContent, "London") {
		t.Errorf("expected log to contain method and city, got: %q", logContent)
	}
}
