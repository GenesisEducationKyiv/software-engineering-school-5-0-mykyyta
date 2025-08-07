package job

import (
	"context"
	"errors"
	"io"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
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
	logger := loggerPkg.From(ctx)

	for {
		task, err := w.queue.Dequeue(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				logger.Info("Worker shutdown signal received")
				return
			case errors.Is(err, io.EOF):
				logger.Info("Queue closed, worker exiting")
				return
			default:
				logger.Error("Failed to dequeue task: %v", err)
				continue
			}
		}

		if task.Email == "" {
			logger.Warn("Empty email in task, skipping")
			continue
		}

		go func(parentCtx context.Context, t Task) {
			taskCtx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
			defer cancel()
			defer func() {
				if r := recover(); r != nil {
					logger := loggerPkg.From(taskCtx)
					logger.Error("Panic recovered while handling task for %s: %v", t.Email, r)
				}
			}()
			logger := loggerPkg.From(taskCtx)
			err := w.subService.ProcessWeatherReportTask(taskCtx, t)
			if err != nil {
				logger.Error("Failed to process task", "email", t.Email, "error", err)
			} else {
				logger.Info("Task processed", "email", t.Email)
			}
		}(ctx, task)
	}
}
