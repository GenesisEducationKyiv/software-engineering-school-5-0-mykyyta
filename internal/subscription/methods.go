package subscription

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

func (s *SubscriptionService) Subscribe(ctx context.Context, email, city, frequency string) error {
	ok, err := s.cityValidator.CityIsValid(ctx, city)
	if err != nil {
		return fmt.Errorf("failed to validate city: %w", err)
	}
	if !ok {
		return ErrCityNotFound
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check existing subscription: %w", err)
	}

	if existing != nil && existing.IsConfirmed && !existing.IsUnsubscribed {
		return ErrEmailAlreadyExists
	}

	token, err := s.tokenService.Generate(email)
	if err != nil {
		return fmt.Errorf("could not generate token: %w", err)
	}

	if existing != nil {
		existing.City = city
		existing.Frequency = frequency
		existing.Token = token
		existing.CreatedAt = time.Now()
		existing.IsConfirmed = false
		existing.IsUnsubscribed = false
		if err := s.repo.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	} else {
		sub := &Subscription{
			ID:             uuid.New().String(),
			Email:          email,
			City:           city,
			Frequency:      frequency,
			Token:          token,
			IsConfirmed:    false,
			IsUnsubscribed: false,
			CreatedAt:      time.Now(),
		}
		if err := s.repo.Create(ctx, sub); err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	go func() {
		if err := s.emailService.SendConfirmationEmail(email, token); err != nil {
			fmt.Printf("Failed to send confirmation email to %s: %v\n", email, err)
		}
	}()

	return nil
}

func (s *SubscriptionService) Confirm(ctx context.Context, token string) error {
	email, err := s.tokenService.Parse(token)
	if err != nil {
		return ErrInvalidToken
	}

	sub, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}
	if sub.IsConfirmed {
		return nil
	}

	sub.IsConfirmed = true
	if err := s.repo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to confirm subscription: %w", err)
	}

	return nil
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	email, err := s.tokenService.Parse(token)
	if err != nil {
		return ErrInvalidToken
	}

	sub, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	if sub.IsUnsubscribed {
		return nil
	}

	sub.IsUnsubscribed = true
	if err := s.repo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	return nil
}

func (s *SubscriptionService) ListConfirmedByFrequency(ctx context.Context, frequency string) ([]Subscription, error) {
	return s.repo.GetConfirmedByFrequency(ctx, frequency)
}
