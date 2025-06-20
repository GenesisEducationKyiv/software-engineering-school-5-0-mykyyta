package jobs

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	"weatherApi/internal/weather"
)

type Task struct {
	Email string
	City  string
	Token string
}

type taskSource interface {
	Dequeue(ctx context.Context) (Task, error)
}

type emailSender interface {
	SendWeatherReport(toEmail string, w weather.Weather, city, token string) error
}

type weatherProvider interface {
	GetWeather(ctx context.Context, city string) (weather.Weather, error)
}

type Worker struct {
	queue          taskSource
	weatherService weatherProvider
	emailService   emailSender
}

func NewWorker(queue taskSource, weather weatherProvider, email emailSender) *Worker {
	return &Worker{
		queue:          queue,
		weatherService: weather,
		emailService:   email,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("[Worker] Started")

	for {
		task, err := w.queue.Dequeue(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				log.Println("[Worker] Shutdown signal received. Stopping...")
				return
			case errors.Is(err, io.EOF):
				log.Println("[Worker] Queue closed. Exiting.")
				return
			default:
				log.Printf("[Worker] Failed to dequeue task: %v", err)
				continue
			}
		}

		if task.Email == "" {
			log.Println("[Worker] Empty email in task. Skipping.")
			continue
		}

		go func(t Task) {
			taskCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Worker] Panic recovered while handling task for %s: %v", t.Email, r)
				}
			}()

			w.handleTask(taskCtx, t)
		}(task)
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
