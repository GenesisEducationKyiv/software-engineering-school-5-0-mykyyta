package sendgrid

import (
	"context"
	"fmt"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

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
	logger := loggerPkg.From(ctx)
	logger.Debugw("Sending email via SendGrid API", "user", loggerPkg.HashEmail(to))

	from := mail.NewEmail("weatherApp", s.from)
	toUser := mail.NewEmail("User", to)

	message := mail.NewSingleEmail(from, subject, toUser, plain, html)

	resp, err := s.client.Send(message)
	if err != nil {
		logger.Errorw("SendGrid API failed",
			"provider", "sendgrid",
			"error", "connection",
			"details", err.Error())
		return fmt.Errorf("sendgrid connection failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		logger.Errorw("SendGrid API error",
			"provider", "sendgrid",
			"status_code", resp.StatusCode,
			"error", "api_response")
		return fmt.Errorf("sendgrid API error: status %d", resp.StatusCode)
	}

	logger.Debugw("Email sent via SendGrid", "status_code", resp.StatusCode)
	return nil
}
