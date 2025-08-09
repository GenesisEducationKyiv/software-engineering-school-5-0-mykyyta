package job

import (
	"context"
	"log"
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
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[Dispatcher] Context cancelled. Stopping dispatcher.")
				return
			case freq, ok := <-d.EventSource.Events():
				if !ok {
					log.Println("[Dispatcher] Event source closed. Exiting.")
					return
				}
				log.Printf("[Dispatcher] Received event: %s", freq)
				d.DispatchScheduledEmails(ctx, freq)
			}
		}
	}()
}

func (d *EmailDispatcher) DispatchScheduledEmails(ctx context.Context, freq string) {
	tasks, err := d.SubService.GenerateWeatherReportTasks(ctx, freq)
	if err != nil {
		log.Printf("[Dispatcher] Failed to generate tasks: %v", err)
		return
	}

	for _, task := range tasks {
		log.Printf("[Dispatcher] Enqueueing task for %q", task.Email)
		if err := d.TaskQueue.Enqueue(ctx, task); err != nil {
			log.Printf("[Dispatcher] Failed to enqueue: %v", err)
		}
	}
}
