package job

import (
	"context"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
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
				logger.Info("Event received", "frequency", freq)
				d.DispatchScheduledEmails(ctx, freq)
			}
		}
	}()
}

func (d *EmailDispatcher) DispatchScheduledEmails(ctx context.Context, freq string) {
	logger := loggerPkg.From(ctx)
	tasks, err := d.SubService.GenerateWeatherReportTasks(ctx, freq)
	if err != nil {
		logger.Error("Failed to generate tasks", "error", err)
		return
	}

	for _, task := range tasks {
		logger.Info("Enqueuing task", "email", task.Email)
		if err := d.TaskQueue.Enqueue(ctx, task); err != nil {
			logger.Error("Failed to enqueue", "error", err)
		}
	}
}
