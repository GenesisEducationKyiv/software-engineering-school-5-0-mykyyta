package subscription

import (
	"context"
	"errors"
)

var (
	ErrCityNotFound         = errors.New("city not found")
	ErrEmailAlreadyExists   = errors.New("email already subscribed")
	ErrInvalidToken         = errors.New("invalid token")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type repo interface {
	GetByEmail(ctx context.Context, email string) (*Subscription, error)
	Create(ctx context.Context, sub *Subscription) error
	Update(ctx context.Context, sub *Subscription) error
	GetConfirmedByFrequency(ctx context.Context, frequency string) ([]Subscription, error)
}

type emailService interface {
	SendConfirmationEmail(email, token string) error
}

type cityValidator interface {
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type tokenService interface {
	Generate(email string) (string, error)
	Parse(tokenStr string) (string, error)
}

type Service struct {
	repo          repo
	emailService  emailService
	cityValidator cityValidator
	tokenService  tokenService
}

func NewService(
	repo repo,
	emailService emailService,
	cityValidator cityValidator,
	tokenService tokenService,
) Service {
	return Service{
		repo:          repo,
		emailService:  emailService,
		cityValidator: cityValidator,
		tokenService:  tokenService,
	}
}
