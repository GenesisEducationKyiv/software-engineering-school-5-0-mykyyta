package jobs

import (
	"context"
	"log"

	"weatherApi/internal/subscription"
)

type TaskQueue interface {
	Enqueue(ctx context.Context, task Task) error
}

type SubService interface {
	ListConfirmedByFrequency(ctx context.Context, frequency string) ([]subscription.Subscription, error)
}

type EmailDispatcher struct {
	SubService SubService
	TaskQueue  TaskQueue
}

func (d *EmailDispatcher) DispatchScheduledEmails(freq string) {
	ctx := context.Background()

	subs, err := d.SubService.ListConfirmedByFrequency(ctx, freq)
	if err != nil {
		log.Printf("[Dispatcher] Failed to fetch subscriptions: %v", err)
		return
	}

	for _, sub := range subs {
		task := Task{
			Email: sub.Email,
			City:  sub.City,
			Token: sub.Token,
		}
		if err := d.TaskQueue.Enqueue(ctx, task); err != nil {
			log.Printf("[Dispatcher] Failed to enqueue: %v", err)
		}
	}
}
