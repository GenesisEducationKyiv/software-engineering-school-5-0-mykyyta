package jobs

import (
	"context"
	"errors"
	"io"
	"log"
	"time"
)

type Task struct {
	Email string
	City  string
	Token string
}

type taskSource interface {
	Dequeue(ctx context.Context) (Task, error)
}

type taskService interface {
	ProcessWeatherReportTask(ctx context.Context, task Task) error
}

type Worker struct {
	queue      taskSource
	subService taskService
}

func NewWorker(queue taskSource, subservice taskService) *Worker {
	return &Worker{
		queue:      queue,
		subService: subservice,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("[Worker] Started")

	for {
		task, err := w.queue.Dequeue(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				log.Println("[Worker] Shutdown signal received. Stopping...")
				return
			case errors.Is(err, io.EOF):
				log.Println("[Worker] Queue closed. Exiting.")
				return
			default:
				log.Printf("[Worker] Failed to dequeue task: %v", err)
				continue
			}
		}

		if task.Email == "" {
			log.Println("[Worker] Empty email in task. Skipping.")
			continue
		}

		go func(t Task) {
			taskCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Worker] Panic recovered while handling task for %s: %v", t.Email, r)
				}
			}()

			err := w.subService.ProcessWeatherReportTask(taskCtx, t)
			if err != nil {
				log.Printf("[Worker] failed to process task for %s: %v", t.Email, err)
			} else {
				log.Printf("[Worker] successfully processed task for %s", t.Email)
			}
		}(task)
	}
}
