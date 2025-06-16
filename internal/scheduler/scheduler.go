package scheduler

import (
	"context"

	"weatherApi/internal/subscription"

	"weatherApi/internal/email"
	"weatherApi/internal/jobs"
	"weatherApi/internal/weather"
)

type WeatherScheduler struct {
	queue      *jobs.LocalQueue
	dispatcher *jobs.EmailDispatcher
	worker     *jobs.Worker
	cron       *jobs.CronScheduler
}

func NewScheduler(
	subService *subscription.SubscriptionService,
	weatherService *weather.WeatherService,
	emailService *email.EmailService,
) *WeatherScheduler {
	queue := jobs.NewLocalQueue(100)

	dispatcher := &jobs.EmailDispatcher{
		SubService: subService,
		TaskQueue:  queue,
	}

	worker := jobs.NewWorker(queue.Channel(), weatherService, emailService)
	cron := jobs.NewCronScheduler(dispatcher.DispatchScheduledEmails)

	return &WeatherScheduler{
		queue:      queue,
		dispatcher: dispatcher,
		worker:     worker,
		cron:       cron,
	}
}

func (s *WeatherScheduler) Start() {
	ctx := context.Background()
	go s.worker.Start(ctx)
	s.cron.Start()
}

func (s *WeatherScheduler) Stop() {
	s.cron.Stop()
	s.queue.Close() // важливо
}
