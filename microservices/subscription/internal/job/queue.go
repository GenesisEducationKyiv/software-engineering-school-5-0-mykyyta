package job

import (
	"context"
	"fmt"
	"io"
	"sync"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type LocalQueue struct {
	queue chan Task
	once  sync.Once
}

func NewLocalQueue(bufferSize int) *LocalQueue {
	return &LocalQueue{
		queue: make(chan Task, bufferSize),
	}
}

func (q *LocalQueue) Enqueue(ctx context.Context, task Task) error {
	logger := loggerPkg.From(ctx)
	if task.Email == "" {
		logger.Warn("Skip enqueue: empty email", "city", task.City)
		return fmt.Errorf("cannot enqueue task: missing email")
	}

	logger.Info("Task enqueued", "email", task.Email)

	select {
	case q.queue <- task:
		return nil
	case <-ctx.Done():
		logger.Warn("Enqueue cancelled by context", "email", task.Email)
		return ctx.Err()
	}
}

func (q *LocalQueue) Dequeue(ctx context.Context) (Task, error) {
	select {
	case task, ok := <-q.queue:
		if !ok {
			return Task{}, io.EOF
		}
		return task, nil
	case <-ctx.Done():
		return Task{}, ctx.Err()
	}
}

func (q *LocalQueue) Close(ctx context.Context) {
	logger := loggerPkg.From(ctx)
	logger.Info("Queue stopping")
	q.once.Do(func() {
		close(q.queue)
	})
	logger.Info("Queue stopped")
}
