package token

type Provider interface {
	Generate(email string) (string, error)
	Parse(token string) (string, error)
}
type Service struct {
	provider Provider
}

func NewService(p Provider) *Service {
	return &Service{provider: p}
}

func (s *Service) Generate(email string) (string, error) {
	return s.provider.Generate(email)
}

func (s *Service) Parse(token string) (string, error) {
	return s.provider.Parse(token)
}
