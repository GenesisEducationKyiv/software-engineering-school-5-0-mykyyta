package email

import (
	"fmt"

	"monolith/internal/domain"
)

type Provider interface {
	Send(to, subject, plainText, html string) error
}

type Service struct {
	provider Provider
	baseURL  string
}

func NewService(provider Provider, baseURL string) Service {
	return Service{
		provider: provider,
		baseURL:  baseURL,
	}
}

func (s Service) SendConfirmationEmail(toEmail, token string) error {
	url := fmt.Sprintf("%s/api/confirm/%s", s.baseURL, token)
	subject := "Підтвердіть вашу підписку"
	plain := "Будь ласка, підтвердіть вашу підписку за цим посиланням: " + url
	html := fmt.Sprintf(`<p>Натисніть <a href="%s">сюди</a> для підтвердження підписки.</p>`, url)

	return s.provider.Send(toEmail, subject, plain, html)
}

func (s Service) SendWeatherReport(toEmail string, weather domain.Report, city, token string) error {
	url := fmt.Sprintf("%s/api/unsubscribe/%s", s.baseURL, token)
	subject := fmt.Sprintf("Оновлення погоди для %s", city)

	plain := fmt.Sprintf(
		"Поточна погода в %s:\nТемпература: %.1f°C\nВологість: %d%%\nОпис: %s\n\nВідписатися: %s",
		city, weather.Temperature, weather.Humidity, weather.Description, url,
	)

	html := fmt.Sprintf(
		`<h2>Погода в %s</h2>
		<p><strong>Температура:</strong> %.1f°C</p>
		<p><strong>Вологість:</strong> %d%%</p>
		<p><strong>Опис:</strong> %s</p>
		<hr>
		<p style="font-size:small">Не хочете більше отримувати ці листи? <a href="%s">Відписатися</a></p>`,
		city, weather.Temperature, weather.Humidity, weather.Description, url,
	)

	return s.provider.Send(toEmail, subject, plain, html)
}
