package email

import (
	"context"

	"email/internal/domain"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
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
	logger.From(ctx).Infow("Starting email sending", "to", req.To, "template", req.Template)

	subject, plain, html, err := s.templateStore.Render(ctx, req.Template, req.Data)
	if err != nil {
		logger.From(ctx).Errorw("Template rendering failed", "template", req.Template, "err", err)
		return err
	}

	if err := s.provider.Send(ctx, req.To, subject, plain, html); err != nil {
		logger.From(ctx).Errorw("Email provider failed", "to", req.To, "err", err)
		return err
	}

	logger.From(ctx).Infow("Email sent successfully", "to", req.To, "template", req.Template)
	return nil
}
