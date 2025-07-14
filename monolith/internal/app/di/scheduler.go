package di

import (
	"context"
	"log"
	"sync"

	job2 "monolith/internal/job"
	"monolith/internal/subscription"
)

type WeatherScheduler struct {
	queue      *job2.LocalQueue
	dispatcher *job2.EmailDispatcher
	worker     *job2.Worker
	cron       *job2.CronEventSource
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewScheduler(
	subService subscription.Service,
) *WeatherScheduler {
	queue := job2.NewLocalQueue(100)
	cron := job2.NewCronEventSource()
	dispatcher := job2.NewEmailDispatcher(subService, queue, cron)
	worker := job2.NewWorker(queue, subService)

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
