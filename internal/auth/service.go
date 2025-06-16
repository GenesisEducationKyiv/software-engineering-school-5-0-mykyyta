package auth

type TokenProvider interface {
	Generate(email string) (string, error)
	Parse(token string) (string, error)
}
type TokenService struct {
	provider TokenProvider
}

func NewTokenService(p TokenProvider) *TokenService {
	return &TokenService{provider: p}
}

func (s *TokenService) Generate(email string) (string, error) {
	return s.provider.Generate(email)
}

func (s *TokenService) Parse(token string) (string, error) {
	return s.provider.Parse(token)
}
