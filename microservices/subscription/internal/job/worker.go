package job

import (
	"context"
	"errors"
	"io"
	"time"

	"subscription/pkg/logger"
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
	lg := logger.From(ctx)

	for {
		task, err := w.queue.Dequeue(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				lg.Info("Worker shutdown signal received")
				return
			case errors.Is(err, io.EOF):
				lg.Info("Queue closed, worker exiting")
				return
			default:
				lg.Errorf("Failed to dequeue task: %v", err)
				continue
			}
		}

		if task.Email == "" {
			lg.Warn("Empty email in task, skipping")
			continue
		}

		go func(parentCtx context.Context, t Task) {
			taskCtx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
			defer cancel()
			defer func() {
				if r := recover(); r != nil {
					lg := logger.From(taskCtx)
					lg.Errorf("Panic recovered while handling task for %s: %v", t.Email, r)
				}
			}()
			lg := logger.From(taskCtx)
			err := w.subService.ProcessWeatherReportTask(taskCtx, t)
			if err != nil {
				lg.Errorf("Failed to process task for %s: %v", t.Email, err)
			} else {
				lg.Infof("Task processed for %s", t.Email)
			}
		}(ctx, task)
	}
}
