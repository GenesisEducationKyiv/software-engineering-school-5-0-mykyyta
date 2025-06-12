package scheduler

import (
	"context"
	"log"

	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"

	"github.com/robfig/cron/v3"
)

type SubscriptionRepository interface {
	ListConfirmedByFrequency(ctx context.Context, frequency string) ([]subscription.Subscription, error)
}

type WeatherFetcher interface {
	GetWeather(ctx context.Context, city string) (*weather.Weather, error)
}

type EmailSender interface {
	SendWeatherReport(toEmail string, w *weather.Weather, city, token string) error
}

type WeatherScheduler struct {
	repo    SubscriptionRepository
	weather WeatherFetcher
	email   EmailSender
}

func NewScheduler(repo SubscriptionRepository, weather WeatherFetcher, email EmailSender) *WeatherScheduler {
	return &WeatherScheduler{
		repo:    repo,
		weather: weather,
		email:   email,
	}
}

func (s *WeatherScheduler) Start() {
	c := cron.New()

	// Щогодини
	_, err := c.AddFunc("@hourly", func() {
		log.Println("[Scheduler] Hourly job triggered")
		s.sendUpdates("hourly")
	})
	if err != nil {
		log.Fatalf("Failed to schedule hourly job: %v", err)
	}

	// Щодня о 12:00 UTC
	_, err = c.AddFunc("0 12 * * *", func() {
		log.Println("[Scheduler] Daily job triggered")
		s.sendUpdates("daily")
	})
	if err != nil {
		log.Fatalf("Failed to schedule daily job: %v", err)
	}

	log.Println("[Scheduler] Started")
	c.Start()
}

func (s *WeatherScheduler) sendUpdates(frequency string) {
	ctx := context.Background()

	subs, err := s.repo.ListConfirmedByFrequency(ctx, frequency)
	if err != nil {
		log.Printf("[Scheduler] Failed to fetch subscriptions: %v", err)
		return
	}

	for _, sub := range subs {
		go func(sub subscription.Subscription) {
			if err := s.processSubscription(ctx, sub); err != nil {
				log.Printf("[Scheduler] Failed to process %s: %v", sub.Email, err)
			} else {
				log.Printf("[Scheduler] Sent weather to %s", sub.Email)
			}
		}(sub)
	}
}

func (s *WeatherScheduler) processSubscription(ctx context.Context, sub subscription.Subscription) error {
	weather, err := s.weather.GetWeather(ctx, sub.City)
	if err != nil {
		return err
	}
	return s.email.SendWeatherReport(sub.Email, weather, sub.City, sub.Token)
}
