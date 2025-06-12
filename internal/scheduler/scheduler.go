package scheduler

import (
	"context"
	"log"
	"time"
	"weatherApi/internal/subscription"

	"weatherApi/internal/email"
	"weatherApi/internal/weather"

	"gorm.io/gorm"
)

// DB is the scheduler's database instance.
// Must be set via SetDB() before StartWeatherScheduler is called.
var (
	DB               *gorm.DB
	FetchWeather     = weather.FetchWithStatus
	SendWeatherEmail = email.SendWeatherEmail
)

// SetDB assigns a GORM database instance to the scheduler.
// This allows decoupling from the main DB package for testability or modularity.
func SetDB(db *gorm.DB) {
	DB = db
}

// StartWeatherScheduler starts a background task that sends weather updates.
// It sends "hourly" updates every round hour and "daily" updates at 12:00 UTC.
func StartWeatherScheduler() {
	log.Println("[Scheduler] started")

	// Align to the next full hour (e.g., xx:00:00)
	now := time.Now()
	nextHour := now.Truncate(time.Hour).Add(time.Hour)
	sleep := time.Until(nextHour)
	log.Printf("[Scheduler] sleeping %v to align to next full hour\n", sleep)
	time.Sleep(sleep)

	ticker := time.NewTicker(1 * time.Hour)
	for {
		now := time.Now()
		log.Println("[Scheduler] running tick", now.Format("15:04:05"))

		go sendWeatherUpdates("hourly")

		if now.Hour() == 12 {
			go sendWeatherUpdates("daily")
		}

		<-ticker.C
	}
}

// sendWeatherUpdates fetches all active subscriptions with the given frequency
// and sends weather updates for each one via email.
func sendWeatherUpdates(frequency string) {
	var subs []subscription.Subscription

	if err := DB.Where(
		"is_confirmed = ? AND is_unsubscribed = ? AND frequency = ?",
		true, false, frequency,
	).Find(&subs).Error; err != nil {
		log.Printf("[Scheduler] Failed to query subscriptions: %v", err)
		return
	}

	for _, sub := range subs {
		if err := ProcessSubscription(context.Background(), sub); err != nil {
			log.Printf("[Scheduler] Failed to process %s: %v", sub.Email, err)
		} else {
			log.Printf("[Scheduler] Weather sent to %s", sub.Email)
		}
	}
}

// ProcessSubscription fetches the weather for a single subscription
// and sends the email using the stored unsubscribe token.
func ProcessSubscription(ctx context.Context, sub subscription.Subscription) error {
	weather, _, err := FetchWeather(ctx, sub.City)
	if err != nil {
		return err
	}
	return SendWeatherEmail(sub.Email, weather, sub.City, sub.Token)
}
