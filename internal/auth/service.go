package auth

type tokenProvider interface {
	Generate(email string) (string, error)
	Parse(token string) (string, error)
}
type TokenService struct {
	provider tokenProvider
}

func NewTokenService(p tokenProvider) *TokenService {
	return &TokenService{provider: p}
}

func (s *TokenService) Generate(email string) (string, error) {
	return s.provider.Generate(email)
}

func (s *TokenService) Parse(token string) (string, error) {
	return s.provider.Parse(token)
}
