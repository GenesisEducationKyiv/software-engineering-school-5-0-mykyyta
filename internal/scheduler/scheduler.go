package scheduler

import (
	"context"
	"log"
	"sync"

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
	wg         sync.WaitGroup
}

func New(
	subService subscription.Service,
	weatherService weather.Service,
	emailService email.Service,
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

	log.Println("[Scheduler] Starting scheduler...")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.worker.Start(ctx)
	}()

	s.dispatcher.Start(ctx)
	s.cron.Start(ctx)
}

func (s *WeatherScheduler) Stop() {
	log.Println("[Scheduler] Stopping scheduler...")

	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()

	s.cron.Stop()
	s.queue.Close()

	log.Println("[Scheduler] Scheduler stopped.")
}
