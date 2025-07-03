package testutils

import (
	"context"
	"weatherApi/internal/domain"
)

// -------------------------------
// FakeTokenProvider
// -------------------------------

type FakeTokenProvider struct {
	FixedEmail string
	ParseErr   error
	GenToken   string
	GenErr     error
}

func (f *FakeTokenProvider) Generate(email string) (string, error) {
	if f.GenErr != nil {
		return "", f.GenErr
	}
	if f.GenToken != "" {
		return f.GenToken, nil
	}
	return "mock-token-" + email, nil
}

func (f *FakeTokenProvider) Parse(token string) (string, error) {
	if f.ParseErr != nil {
		return "", f.ParseErr
	}
	if f.FixedEmail != "" {
		return f.FixedEmail, nil
	}
	return "test@example.com", nil
}

// -------------------------------
// FakeEmailProvider
// -------------------------------

type FakeEmailProvider struct {
	To      string
	Subject string
	Plain   string
	HTML    string
	Sent    bool
	Err     error
}

func (f *FakeEmailProvider) Send(to, subject, plain, html string) error {
	if f.Err != nil {
		return f.Err
	}
	f.To = to
	f.Subject = subject
	f.Plain = plain
	f.HTML = html
	f.Sent = true
	return nil
}

// -------------------------------
// FakeWeatherProvider
// -------------------------------

type FakeWeatherProvider struct {
	Valid         bool
	CityExistsErr error
	Weather       domain.Report
	WeatherSet    bool
	WeatherErr    error
}

func (f *FakeWeatherProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	if f.CityExistsErr != nil {
		return false, f.CityExistsErr
	}
	return f.Valid, nil
}

func (f *FakeWeatherProvider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	if f.WeatherErr != nil {
		return domain.Report{}, f.WeatherErr
	}
	if f.WeatherSet {
		return f.Weather, nil
	}
	return domain.Report{
		Temperature: 22.5,
		Humidity:    55,
		Description: "Sunny",
	}, nil
}
