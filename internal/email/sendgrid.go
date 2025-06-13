package email

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridProvider struct {
	apiKey string
	from   string
}

func NewSendGridProvider(apiKey, from string) *SendGridProvider {
	return &SendGridProvider{
		apiKey: apiKey,
		from:   from,
	}
}

func (s *SendGridProvider) Send(toEmail, subject, plainTextContent, htmlContent string) error {
	from := mail.NewEmail("weatherApp", s.from)
	to := mail.NewEmail("User", toEmail)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	client := sendgrid.NewSendClient(s.apiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("SendGrid failed with status %d: %s", response.StatusCode, response.Body)
	}

	return nil
}
