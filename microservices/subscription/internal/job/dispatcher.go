package job

import (
	"context"

	loggerPkg "subscription/pkg/logger"
)

type eventSource interface {
	Events() <-chan string
}

type taskQueue interface {
	Enqueue(ctx context.Context, task Task) error
}

type subservice interface {
	GenerateWeatherReportTasks(ctx context.Context, frequency string) ([]Task, error)
}

type EmailDispatcher struct {
	SubService  subservice
	TaskQueue   taskQueue
	EventSource eventSource
}

func NewEmailDispatcher(subService subservice, taskQueue taskQueue, eventSource eventSource) *EmailDispatcher {
	return &EmailDispatcher{
		SubService:  subService,
		TaskQueue:   taskQueue,
		EventSource: eventSource,
	}
}

func (d *EmailDispatcher) Start(ctx context.Context) {
	logger := loggerPkg.From(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("Dispatcher context cancelled, stopping")
				return
			case freq, ok := <-d.EventSource.Events():
				if !ok {
					logger.Info("Event source closed, dispatcher exiting")
					return
				}
				logger.Infof("Event received: %s", freq)
				d.DispatchScheduledEmails(ctx, freq)
			}
		}
	}()
}

func (d *EmailDispatcher) DispatchScheduledEmails(ctx context.Context, freq string) {
	logger := loggerPkg.From(ctx)
	tasks, err := d.SubService.GenerateWeatherReportTasks(ctx, freq)
	if err != nil {
		logger.Errorf("Failed to generate tasks: %v", err)
		return
	}

	for _, task := range tasks {
		logger.Infof("Enqueuing task for %q", task.Email)
		if err := d.TaskQueue.Enqueue(ctx, task); err != nil {
			logger.Errorf("Failed to enqueue: %v", err)
		}
	}
}
