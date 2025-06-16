package jobs

import (
	"context"
	"log"
	"time"

	"weatherApi/internal/weather"
)

type EmailSender interface {
	SendWeatherReport(toEmail string, w weather.Weather, city, token string) error
}

type WeatherProvider interface {
	GetWeather(ctx context.Context, city string) (weather.Weather, error)
}

type Worker struct {
	queue          <-chan Task
	weatherService WeatherProvider
	emailService   EmailSender
}

func NewWorker(queue <-chan Task, weather WeatherProvider, email EmailSender) *Worker {
	return &Worker{
		queue:          queue,
		weatherService: weather,
		emailService:   email,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("[Worker] Started")

	for {
		select {
		case task := <-w.queue:
			go func(t Task) {
				taskCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()
				w.handleTask(taskCtx, t)
			}(task)

		case <-ctx.Done():
			log.Println("[Worker] Context cancelled. Exiting.")
			return
		}
	}
}

func (w *Worker) handleTask(ctx context.Context, t Task) {
	log.Printf("[Worker] Processing task for %s", t.Email)

	weather, err := w.weatherService.GetWeather(ctx, t.City)
	if err != nil {
		log.Printf("[Worker] Failed to get weather for %s: %v", t.City, err)
		return
	}

	err = w.emailService.SendWeatherReport(t.Email, weather, t.City, t.Token)
	if err != nil {
		log.Printf("[Worker] Failed to send email to %s: %v", t.Email, err)
	} else {
		log.Printf("[Worker] Email sent to %s", t.Email)
	}
}
