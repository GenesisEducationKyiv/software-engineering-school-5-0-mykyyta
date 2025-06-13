package auth

type TokenProvider interface {
	Generate(email string) (string, error)
	Parse(token string) (string, error)
}
type TokenService struct {
	provider TokenProvider
}

// NewTokenService створює сервіс з будь-яким постачальником токенів.
func NewTokenService(p TokenProvider) *TokenService {
	return &TokenService{provider: p}
}

// Generate делегує генерацію токена провайдеру.
func (s *TokenService) Generate(email string) (string, error) {
	return s.provider.Generate(email)
}

// Parse делегує парсинг токена провайдеру.
func (s *TokenService) Parse(token string) (string, error) {
	return s.provider.Parse(token)
}
