package email_test

import (
	"fmt"
	"testing"
	"weatherApi/monolith/internal/domain"
	"weatherApi/monolith/internal/email"

	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	lastTo      string
	lastSubject string
	lastPlain   string
	lastHTML    string
}

func (m *mockProvider) Send(to, subject, plainText, html string) error {
	m.lastTo = to
	m.lastSubject = subject
	m.lastPlain = plainText
	m.lastHTML = html
	return nil
}

func TestSendWeatherReport_EdgeCases_UkrainianTemplate(t *testing.T) {
	tests := []struct {
		name     string
		city     string
		token    string
		weather  domain.Report
		expected map[string]string
	}{
		{
			name:  "Normal case, English city and description",
			city:  "London",
			token: "tok123",
			weather: domain.Report{
				Temperature: 21.5,
				Humidity:    65,
				Description: "clear sky",
			},
		},
		{
			name:  "Negative temperature",
			city:  "Oslo",
			token: "tok_snow",
			weather: domain.Report{
				Temperature: -15.2,
				Humidity:    80,
				Description: "snow",
			},
		},
		{
			name:  "Empty city, empty token",
			city:  "",
			token: "",
			weather: domain.Report{
				Temperature: 5.0,
				Humidity:    40,
				Description: "fog",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProvider{}
			service := email.NewService(mock, "https://example.com")

			err := service.SendWeatherReport("test@example.com", tt.weather, tt.city, tt.token)
			assert.NoError(t, err)

			// 🔍 Check plain
			assert.Contains(t, mock.lastPlain, fmt.Sprintf("Поточна погода в %s", tt.city))
			assert.Contains(t, mock.lastPlain, fmt.Sprintf("Температура: %.1f°C", tt.weather.Temperature))
			assert.Contains(t, mock.lastPlain, fmt.Sprintf("Вологість: %d%%", tt.weather.Humidity))
			assert.Contains(t, mock.lastPlain, fmt.Sprintf("Опис: %s", tt.weather.Description))

			// 🔍 Check HTML
			assert.Contains(t, mock.lastHTML, fmt.Sprintf("<h2>Погода в %s</h2>", tt.city))
			assert.Contains(t, mock.lastHTML, fmt.Sprintf("Температура:</strong> %.1f°C", tt.weather.Temperature))
			assert.Contains(t, mock.lastHTML, fmt.Sprintf("Вологість:</strong> %d%%", tt.weather.Humidity))
			assert.Contains(t, mock.lastHTML, fmt.Sprintf("Опис:</strong> %s", tt.weather.Description))
			assert.Contains(t, mock.lastHTML, fmt.Sprintf("/api/unsubscribe/%s", tt.token))
		})
	}
}
