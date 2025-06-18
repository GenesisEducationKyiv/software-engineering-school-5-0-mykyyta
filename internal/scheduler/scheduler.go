package scheduler

import (
	"context"

	"weatherApi/internal/email"
	"weatherApi/internal/jobs"
	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"
)

type WeatherScheduler struct {
	queue      *jobs.LocalQueue
	dispatcher *jobs.EmailDispatcher
	worker     *jobs.Worker
	cron       *jobs.CronEventSource
	cancel     context.CancelFunc
}

func NewScheduler(
	subService *subscription.SubscriptionService,
	weatherService *weather.WeatherService,
	emailService *email.EmailService,
) *WeatherScheduler {

	queue := jobs.NewLocalQueue(100)
	cron := jobs.NewCronEventSource()
	dispatcher := jobs.NewEmailDispatcher(subService, queue, cron)
	worker := jobs.NewWorker(queue, weatherService, emailService)

	return &WeatherScheduler{
		queue:      queue,
		dispatcher: dispatcher,
		worker:     worker,
		cron:       cron,
	}
}

func (s *WeatherScheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.cron.Start()
	s.dispatcher.Start(ctx)
	go s.worker.Start(ctx)
}

func (s *WeatherScheduler) Stop() {
	s.cron.Stop()
	s.queue.Close()

	if s.cancel != nil {
		s.cancel()
	}
}
