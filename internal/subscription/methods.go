package subscription

import (
	"errors"
	"fmt"
	"time"

	"weatherApi/internal/jobs"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

func (s Service) Subscribe(ctx context.Context, email, city, frequency string) error {
	_, err := s.weatherService.CityIsValid(ctx, city)
	if err != nil {
		if errors.Is(err, ErrCityNotFound) {
			return ErrCityNotFound
		}
		return fmt.Errorf("failed to validate city: %w", err)
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
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

	if err := s.emailService.SendConfirmationEmail(email, token); err != nil {
		fmt.Printf("Failed to send confirmation email to %s: %v\n", email, err)
	}

	return nil
}

func (s Service) Confirm(ctx context.Context, token string) error {
	email, err := s.tokenService.Parse(token)
	if err != nil {
		return ErrInvalidToken
	}

	sub, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrSubscriptionNotFound) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("failed to get subscription: %w", err)
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

func (s Service) Unsubscribe(ctx context.Context, token string) error {
	email, err := s.tokenService.Parse(token)
	if err != nil {
		return ErrInvalidToken
	}

	sub, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
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

func (s Service) GenerateWeatherReportTasks(ctx context.Context, frequency string) ([]jobs.Task, error) {
	subs, err := s.ListConfirmedByFrequency(ctx, frequency)
	if err != nil {
		return nil, err
	}

	tasks := make([]jobs.Task, 0, len(subs))
	for _, sub := range subs {
		tasks = append(tasks, jobs.Task{
			Email: sub.Email,
			City:  sub.City,
			Token: sub.Token,
		})
	}
	return tasks, nil
}

func (s Service) ListConfirmedByFrequency(ctx context.Context, frequency string) ([]Subscription, error) {
	return s.repo.GetConfirmedByFrequency(ctx, frequency)
}

func (s Service) ProcessWeatherReportTask(ctx context.Context, task jobs.Task) error {
	report, err := s.weatherService.GetWeather(ctx, task.City)
	if err != nil {
		return fmt.Errorf("get weather for %s: %w", task.City, err)
	}

	if err := s.emailService.SendWeatherReport(task.Email, report, task.City, task.Token); err != nil {
		return fmt.Errorf("send email to %s: %w", task.Email, err)
	}

	return nil
}
