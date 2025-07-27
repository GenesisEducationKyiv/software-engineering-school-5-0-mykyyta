package email

import (
	"context"
	"fmt"

	"email/internal/domain"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type Provider interface {
	Send(ctx context.Context, to, subject, plain, html string) error
}

type TemplateRenderer interface {
	Render(ctx context.Context, template domain.TemplateName, data map[string]string) (subject, plain, html string, err error)
}

type Service struct {
	provider      Provider
	templateStore TemplateRenderer
}

func NewService(provider Provider, templateStore TemplateRenderer) Service {
	return Service{
		provider:      provider,
		templateStore: templateStore,
	}
}

func (s Service) Send(ctx context.Context, req domain.SendEmailRequest) error {
	logger := loggerPkg.From(ctx)
	logger.Debugw("Starting email send operation",
		"user", loggerPkg.HashEmail(req.To),
		"template", string(req.Template),
		"data_fields", len(req.Data))

	subject, plain, html, err := s.templateStore.Render(ctx, req.Template, req.Data)
	if err != nil {
		logger.Errorw("Template processing failed",
			"user", loggerPkg.HashEmail(req.To),
			"template", string(req.Template))
		return fmt.Errorf("template processing failed: %w", err)
	}

	if err := s.provider.Send(ctx, req.To, subject, plain, html); err != nil {
		logger.Errorw("Email delivery failed",
			"user", loggerPkg.HashEmail(req.To),
			"template", string(req.Template))
		return fmt.Errorf("email delivery failed: %w", err)
	}

	logger.Debugw("Email sent successfully",
		"user", loggerPkg.HashEmail(req.To),
		"template", string(req.Template))
	return nil
}
