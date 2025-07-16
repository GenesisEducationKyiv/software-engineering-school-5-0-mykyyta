package email

import "email/internal/domain"

type Sender interface {
	Send(req domain.SendEmailRequest) error
}

type Provider interface {
	Send(to, subject, plain, html string) error
}

type TemplateRenderer interface {
	Render(template domain.TemplateName, data map[string]string) (subject, plain, html string, err error)
}

type Service struct {
	provider      Provider
	templateStore TemplateRenderer
}

func NewService(provider Provider, templateStore TemplateRenderer) Sender {
	return &Service{
		provider:      provider,
		templateStore: templateStore,
	}
}

func (s *Service) Send(req domain.SendEmailRequest) error {
	subject, plain, html, err := s.templateStore.Render(req.Template, req.Data)
	if err != nil {
		return err
	}

	return s.provider.Send(req.To, subject, plain, html)
}
