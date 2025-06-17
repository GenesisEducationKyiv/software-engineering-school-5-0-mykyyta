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
	Dispatcher *jobs.EmailDispatcher
	worker     *jobs.Worker
	cron       *jobs.CronScheduler
	cancel     context.CancelFunc
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
		Dispatcher: dispatcher,
		worker:     worker,
		cron:       cron,
	}
}

func (s *WeatherScheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.worker.Start(ctx)
	s.cron.Start()
}

func (s *WeatherScheduler) Stop() {
	s.cron.Stop()
	s.queue.Close()
	if s.cancel != nil {
		s.cancel() // Завершує воркер
	}
}
