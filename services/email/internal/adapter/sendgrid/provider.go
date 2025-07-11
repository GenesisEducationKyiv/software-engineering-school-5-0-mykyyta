package sendgrid

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGrid struct {
	client *sendgrid.Client
	from   string
}

func New(apiKey, from string) *SendGrid {
	return &SendGrid{
		client: sendgrid.NewSendClient(apiKey),
		from:   from,
	}
}

func (s *SendGrid) Send(to, subject, plain, html string) error {
	from := mail.NewEmail("weatherApp", s.from)
	toUser := mail.NewEmail("User", to)

	message := mail.NewSingleEmail(from, subject, toUser, plain, html)

	resp, err := s.client.Send(message)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: status %d - %s", resp.StatusCode, resp.Body)
	}
	return nil
}
