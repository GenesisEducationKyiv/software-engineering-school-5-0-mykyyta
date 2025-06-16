package jobs

import (
	"context"
	"sync"
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
	select {
	case q.queue <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *LocalQueue) Channel() <-chan Task {
	return q.queue
}

// Close closes the task queue channel safely.
func (q *LocalQueue) Close() {
	q.once.Do(func() {
		close(q.queue)
	})
}
