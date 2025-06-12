package subscription

import (
	"context"
	"fmt"
	"time"

	"weatherApi/internal/jwtutil"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo          subscriptionRepository
	emailSender   emailSender
	cityValidator cityValidator
}

type subscriptionRepository interface {
	GetByEmail(ctx context.Context, email string) (*Subscription, error)
	Create(ctx context.Context, sub *Subscription) error
	Update(ctx context.Context, sub *Subscription) error
}

type emailSender interface {
	SendConfirmationEmail(email, token string) error
}

type cityValidator interface {
	IsValid(ctx context.Context, city string) (bool, error)
}

func NewSubscriptionService(
	repo subscriptionRepository,
	emailSender emailSender,
	cityValidator cityValidator,
) *SubscriptionService {
	return &SubscriptionService{
		repo:          repo,
		emailSender:   emailSender,
		cityValidator: cityValidator,
	}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, email, city, frequency string) error {
	ok, err := s.cityValidator.IsValid(ctx, city)
	if err != nil {
		return fmt.Errorf("failed to validate city: %w", err)
	}
	if !ok {
		return fmt.Errorf("city not found")
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check existing subscription: %w", err)
	}

	token, err := jwtutil.Generate(email)
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
		if err := s.emailSender.SendConfirmationEmail(email, token); err != nil {
			fmt.Printf("Failed to send email to %s: %v\n", email, err)
		}
	}()

	return nil
}

func (s *SubscriptionService) Confirm(ctx context.Context, token string) error {
	email, err := jwtutil.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	sub, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("subscription not found")
	}
	if sub.IsConfirmed {
		return nil
	}

	sub.IsConfirmed = true
	if err := s.repo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to confirm subscription: %w", err)
	}

	//if err := scheduler.ProcessSubscription(ctx, *sub); err != nil {
	//	return fmt.Errorf("failed to send forecast: %w", err)
	//}

	return nil
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	email, err := jwtutil.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	sub, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("subscription not found")
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
