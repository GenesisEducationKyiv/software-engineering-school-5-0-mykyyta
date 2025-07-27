package di

import (
	"context"
	"sync"

	"subscription/internal/job"
	"subscription/internal/subscription"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
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

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		wLogger := logger.From(ctx).With("module", "scheduler")
		s.worker.Start(logger.With(ctx, wLogger))
	}()

	dLogger := logger.From(ctx).With("module", "scheduler")
	s.dispatcher.Start(logger.With(ctx, dLogger))

	cLogger := logger.From(ctx).With("module", "scheduler")
	s.cron.Start(logger.With(ctx, cLogger))
}

func (s *WeatherScheduler) Stop(ctx context.Context) {
	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()

	cLogger := logger.From(ctx).With("module", "scheduler")
	s.cron.Stop(logger.With(ctx, cLogger))

	qLogger := logger.From(ctx).With("module", "scheduler")
	s.queue.Close(logger.With(ctx, qLogger))
}
