package scheduler

import (
	"context"
	"log"

	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"

	"github.com/robfig/cron/v3"
)

type subService interface {
	ListConfirmedByFrequency(ctx context.Context, frequency string) ([]subscription.Subscription, error)
}

type weatherService interface {
	GetWeather(ctx context.Context, city string) (weather.Weather, error)
}

type emailService interface {
	SendWeatherReport(toEmail string, w weather.Weather, city, token string) error
}

type WeatherScheduler struct {
	subService     subService
	weatherService weatherService
	emailService   emailService
	cron           *cron.Cron
}

func NewScheduler(sub subService, weather weatherService, email emailService) *WeatherScheduler {
	return &WeatherScheduler{
		subService:     sub,
		weatherService: weather,
		emailService:   email,
		cron:           cron.New(),
	}
}

func (s *WeatherScheduler) Start() {
	_, err := s.cron.AddFunc("0 * * * *", func() {
		log.Println("[Scheduler] Hourly job triggered on the dot")
		s.sendUpdates("hourly")
	})
	if err != nil {
		log.Fatalf("Failed to schedule hourly job: %v", err)
	}

	_, err = s.cron.AddFunc("0 12 * * *", func() {
		log.Println("[Scheduler] Daily job triggered")
		s.sendUpdates("daily")
	})
	if err != nil {
		log.Fatalf("Failed to schedule daily job: %v", err)
	}

	log.Println("[Scheduler] Started")
	s.cron.Start()
}

func (s *WeatherScheduler) Stop() {
	log.Println("[Scheduler] Stopping...")
	s.cron.Stop()
	log.Println("[Scheduler] Stopped")
}

func (s *WeatherScheduler) sendUpdates(frequency string) {
	ctx := context.Background()

	subs, err := s.subService.ListConfirmedByFrequency(ctx, frequency)
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
	weather, err := s.weatherService.GetWeather(ctx, sub.City)
	if err != nil {
		return err
	}
	return s.emailService.SendWeatherReport(sub.Email, weather, sub.City, sub.Token)
}
