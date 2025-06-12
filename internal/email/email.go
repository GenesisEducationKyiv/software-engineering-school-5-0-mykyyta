package email

import (
	"fmt"
	"weatherApi/internal/weather"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"weatherApi/config"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendEmail sends an email via SendGrid using environment variables:
// - SENDGRID_API_KEY: API key for authentication
// - EMAIL_FROM: sender email address
// Fails if SendGrid responds with status code >= 400.
func SendEmail(toEmail, subject, plainTextContent, htmlContent string) error {
	from := mail.NewEmail("weatherApp", config.C.EmailFrom)
	to := mail.NewEmail("User", toEmail)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	client := sendgrid.NewSendClient(config.C.SendGridKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("SendGrid failed with status %d: %s", response.StatusCode, response.Body)
	}

	return nil
}

// SendConfirmationEmail sends a confirmation link to the user's email.
// The token is embedded as part of a URL and used for verifying the subscription.
func SendConfirmationEmail(toEmail, token string) error {
	subject := "Підтвердіть вашу підписку на погодні сповіщення"

	confirmURL := fmt.Sprintf("%s/api/confirm/%s", config.C.BaseURL, token)
	plainText := "Будь ласка, підтвердіть вашу підписку: " + confirmURL
	htmlContent := fmt.Sprintf(
		`<p>Натисніть нижче для підтвердження вашої підписки:</p><p><a href="%s">Підтвердити підписку</a></p>`,
		confirmURL,
	)

	return SendEmail(toEmail, subject, plainText, htmlContent)
}

// SendWeatherEmail sends a weather update to the user with an unsubscribe link.
// The token is used in the unsubscribe URL and must be securely generated.
func SendWeatherEmail(toEmail string, weather *weather.Weather, city, token string) error {
	caser := cases.Title(language.English)
	subject := fmt.Sprintf("Ваше оновлення погоди для %s", caser.String(city))

	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", config.C.BaseURL, token)

	plainText := fmt.Sprintf(
		"Вітаємо!\n\nПоточна погода в %s:\nТемпература: %.1f°C\nВологість: %d%%\nОпис: %s\n\nЯкщо бажаєте скасувати підписку, перейдіть за посиланням: %s",
		caser.String(city), weather.Temperature, weather.Humidity, weather.Description, unsubscribeURL,
	)

	htmlContent := fmt.Sprintf(
		`<h2>Погода в %s</h2>
		<p><strong>Температура:</strong> %.1f°C</p>
		<p><strong>Вологість:</strong> %d%%</p>
		<p><strong>Опис:</strong> %s</p>
		<hr>
		<p style="font-size:small">Не хочете більше отримувати? <a href="%s">Відписатися</a></p>`,
		caser.String(city), weather.Temperature, weather.Humidity, weather.Description, unsubscribeURL,
	)

	return SendEmail(toEmail, subject, plainText, htmlContent)
}
