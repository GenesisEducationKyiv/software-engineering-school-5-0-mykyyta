package jobs

import (
	"context"
	"fmt"
	"io"
	"log"
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
	if task.Email == "" {
		log.Printf("[Queue] Skip enqueue: empty email (city=%q)", task.City)
		return fmt.Errorf("cannot enqueue task: missing email")
	}

	log.Printf("[Queue] Enqueuing task for: %q", task.Email)

	select {
	case q.queue <- task:
		return nil
	case <-ctx.Done():
		log.Printf("[Queue] Context cancelled while enqueuing for: %q", task.Email)
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

func (q *LocalQueue) Close() {
	q.once.Do(func() {
		close(q.queue)
	})
}
