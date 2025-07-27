package sendgrid

import (
	"context"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

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

func (s *SendGrid) Send(ctx context.Context, to, subject, plain, html string) error {
	logger.From(ctx).Infow("Sending email via SendGrid API", "to", to, "from", s.from)

	from := mail.NewEmail("weatherApp", s.from)
	toUser := mail.NewEmail("User", to)

	message := mail.NewSingleEmail(from, subject, toUser, plain, html)

	resp, err := s.client.Send(message)
	if err != nil {
		logger.From(ctx).Errorw("SendGrid API request failed", "to", to, "err", err)
		return err
	}

	if resp.StatusCode >= 400 {
		logger.From(ctx).Errorw("SendGrid API error response",
			"to", to,
			"status_code", resp.StatusCode,
			"response_body", resp.Body)
		return fmt.Errorf("sendgrid error: status %d - %s", resp.StatusCode, resp.Body)
	}

	logger.From(ctx).Infow("Email sent successfully via SendGrid",
		"to", to,
		"status_code", resp.StatusCode)
	return nil
}
