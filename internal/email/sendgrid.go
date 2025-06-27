package email

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGrid struct {
	apiKey string
	from   string
}

func NewSendgrid(apiKey, from string) *SendGrid {
	return &SendGrid{
		apiKey: apiKey,
		from:   from,
	}
}

func (s *SendGrid) Send(toEmail, subject, plainTextContent, htmlContent string) error {
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
