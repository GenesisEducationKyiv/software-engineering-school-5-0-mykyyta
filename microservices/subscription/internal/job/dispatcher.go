package job

import (
	"context"

	"subscription/pkg/logger"
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
	lg := logger.From(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				lg.Info("Dispatcher context cancelled, stopping")
				return
			case freq, ok := <-d.EventSource.Events():
				if !ok {
					lg.Info("Event source closed, dispatcher exiting")
					return
				}
				lg.Infof("Event received: %s", freq)
				d.DispatchScheduledEmails(ctx, freq)
			}
		}
	}()
}

func (d *EmailDispatcher) DispatchScheduledEmails(ctx context.Context, freq string) {
	lg := logger.From(ctx)
	tasks, err := d.SubService.GenerateWeatherReportTasks(ctx, freq)
	if err != nil {
		lg.Errorf("Failed to generate tasks: %v", err)
		return
	}

	for _, task := range tasks {
		lg.Infof("Enqueuing task for %q", task.Email)
		if err := d.TaskQueue.Enqueue(ctx, task); err != nil {
			lg.Errorf("Failed to enqueue: %v", err)
		}
	}
}
