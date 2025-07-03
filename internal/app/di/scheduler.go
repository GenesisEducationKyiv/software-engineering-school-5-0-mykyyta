package di

import (
	"context"
	"log"
	"sync"

	"weatherApi/internal/job"
	"weatherApi/internal/subscription"
)

type WeatherScheduler struct {
	queue      *job.LocalQueue
	dispatcher *job.EmailDispatcher
	worker     *job.Worker
	cron       *job.CronEventSource
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewScheduler(
	subService subscription.Service,
) *WeatherScheduler {
	queue := job.NewLocalQueue(100)
	cron := job.NewCronEventSource()
	dispatcher := job.NewEmailDispatcher(subService, queue, cron)
	worker := job.NewWorker(queue, subService)

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
