package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"subscription/internal/domain"
	"subscription/internal/job"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"github.com/google/uuid"
)

var (
	ErrCityNotFound         = errors.New("city not found")
	ErrEmailAlreadyExists   = errors.New("email already subscribed")
	ErrInvalidToken         = errors.New("invalid token")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type repo interface {
	GetByEmail(ctx context.Context, email string) (*domain.Subscription, error)
	Create(ctx context.Context, sub *domain.Subscription) error
	Update(ctx context.Context, sub *domain.Subscription) error
	GetConfirmedByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error)
}

type emailClient interface {
	SendConfirmationEmail(ctx context.Context, email, token string, idKey string) error
	SendWeatherReport(ctx context.Context, email string, weather domain.Report, city, token string, idKey string) error
}

type WeatherClient interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type tokenService interface {
	Generate(email string) (string, error)
	Parse(tokenStr string) (string, error)
}

type Service struct {
	repo           repo
	emailService   emailClient
	weatherService WeatherClient
	tokenService   tokenService
}

func NewService(
	repo repo,
	emailService emailClient,
	weatherService WeatherClient,
	tokenService tokenService,
) Service {
	return Service{
		repo:           repo,
		emailService:   emailService,
		weatherService: weatherService,
		tokenService:   tokenService,
	}
}

func (s Service) Subscribe(ctx context.Context, email, city string, frequency domain.Frequency) error {
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

	if err := s.createOrUpdateSubscription(ctx, existing, email, city, frequency, token); err != nil {
		return err
	}

	idKey := s.generateIdempotencyKey(email, token)
	if err := s.emailService.SendConfirmationEmail(ctx, email, token, idKey); err != nil {
		logger := loggerPkg.From(ctx)
		logger.Error("Failed to send confirmation email to %s: %v", email, err)
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

func (s Service) GenerateWeatherReportTasks(ctx context.Context, frequency string) ([]job.Task, error) {
	subs, err := s.listConfirmedByFrequency(ctx, frequency)
	if err != nil {
		return nil, err
	}

	tasks := make([]job.Task, 0, len(subs))
	for _, sub := range subs {
		tasks = append(tasks, job.Task{
			Email: sub.Email,
			City:  sub.City,
			Token: sub.Token,
		})
	}
	return tasks, nil
}

func (s Service) ProcessWeatherReportTask(ctx context.Context, task job.Task) error {
	report, err := s.weatherService.GetWeather(ctx, task.City)
	if err != nil {
		return fmt.Errorf("get weather for %s: %w", task.City, err)
	}

	nowHour := time.Now().UTC().Format("2006-01-02T15")
	idKey := fmt.Sprintf("report:%s:%s", task.Email, nowHour)
	if err := s.emailService.SendWeatherReport(ctx, task.Email, report, task.City, task.Token, idKey); err != nil {
		return fmt.Errorf("send email to %s: %w", task.Email, err)
	}

	return nil
}

func (s Service) listConfirmedByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error) {
	return s.repo.GetConfirmedByFrequency(ctx, frequency)
}

func (s Service) createOrUpdateSubscription(ctx context.Context, existing *domain.Subscription, email, city string, frequency domain.Frequency, token string) error {
	now := time.Now()

	if existing != nil {
		updatedSub := &domain.Subscription{
			ID:             existing.ID,
			Email:          existing.Email,
			City:           city,
			Frequency:      frequency,
			Token:          token,
			IsConfirmed:    false,
			IsUnsubscribed: false,
			CreatedAt:      now,
		}
		if err := s.repo.Update(ctx, updatedSub); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	} else {
		sub := &domain.Subscription{
			ID:             uuid.New().String(),
			Email:          email,
			City:           city,
			Frequency:      frequency,
			Token:          token,
			IsConfirmed:    false,
			IsUnsubscribed: false,
			CreatedAt:      now,
		}
		if err := s.repo.Create(ctx, sub); err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	return nil
}

func (s Service) generateIdempotencyKey(email, token string) string {
	return fmt.Sprintf("confirm:%s:%s", email, token)
}
