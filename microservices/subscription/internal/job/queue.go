package job

import (
	"context"
	"fmt"
	"io"
	"sync"

	"subscription/pkg/logger"
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
	lg := logger.From(ctx)
	if task.Email == "" {
		lg.Warnw("Skip enqueue: empty email", "city", task.City)
		return fmt.Errorf("cannot enqueue task: missing email")
	}

	lg.Infow("Task enqueued", "email", task.Email)

	select {
	case q.queue <- task:
		return nil
	case <-ctx.Done():
		lg.Warnw("Enqueue cancelled by context", "email", task.Email)
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
	lg := logger.From(ctx)
	lg.Info("Queue stopping")
	q.once.Do(func() {
		close(q.queue)
	})
	lg.Info("Queue stopped")
}
