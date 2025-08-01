package di

import (
	"context"
	"sync"

	"subscription/internal/job"
	"subscription/internal/subscription"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
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
		wLogger := loggerPkg.From(ctx).With("module", "scheduler")
		s.worker.Start(loggerPkg.With(ctx, wLogger))
	}()

	dLogger := loggerPkg.From(ctx).With("module", "scheduler")
	s.dispatcher.Start(loggerPkg.With(ctx, dLogger))

	cLogger := loggerPkg.From(ctx).With("module", "scheduler")
	s.cron.Start(loggerPkg.With(ctx, cLogger))
}

func (s *WeatherScheduler) Stop(ctx context.Context) {
	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()

	cLogger := loggerPkg.From(ctx).With("module", "scheduler")
	s.cron.Stop(loggerPkg.With(ctx, cLogger))

	qLogger := loggerPkg.From(ctx).With("module", "scheduler")
	s.queue.Close(loggerPkg.With(ctx, qLogger))
}
