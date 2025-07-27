package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"email/internal/domain"
	"email/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageSource interface {
	Consume(ctx context.Context) (<-chan amqp.Delivery, error)
}

type IdempotencyStore interface {
	IsProcessed(ctx context.Context, messageID string) (bool, error)
	MarkAsProcessing(ctx context.Context, messageID string) (bool, error)
	MarkAsProcessed(ctx context.Context, messageID string) error
	ClearProcessing(ctx context.Context, messageID string) error
}

type CircuitBreaker interface {
	CanExecute() bool
	RecordSuccess()
	RecordFailure()
}

type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type EmailUseCase interface {
	Send(req domain.SendEmailRequest) error
}
type Consumer struct {
	source      MessageSource
	useCase     EmailUseCase
	idempotency IdempotencyStore
	breaker     CircuitBreaker
}

func NewConsumer(source MessageSource, useCase EmailUseCase, idempotency IdempotencyStore, breaker CircuitBreaker) *Consumer {
	return &Consumer{
		source:      source,
		useCase:     useCase,
		idempotency: idempotency,
		breaker:     breaker,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.source.Consume(ctx)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	logger.From(ctx).Infow("Consumer started")

	for {
		select {
		case <-ctx.Done():
			logger.From(ctx).Infow("Consumer stopped")
			return ctx.Err()

		case msg, ok := <-msgs:
			if !ok {
				logger.From(ctx).Infow("Message channel closed")
				return nil
			}

			c.handle(ctx, msg)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, msg amqp.Delivery) {
	id := msg.MessageId
	if id == "" {
		logger.From(ctx).Warnw("Skipping message without ID", "rk", msg.RoutingKey)
		_ = msg.Nack(false, false)
		return
	}

	if !c.breaker.CanExecute() {
		logger.From(ctx).Warnw("Circuit breaker is open, rejecting message", "id", id)
		delay := time.Duration(500+rand.Intn(500)) * time.Millisecond
		time.Sleep(delay)
		_ = msg.Nack(false, true)
		return
	}

	err := c.processIdempotently(ctx, id, msg)
	if err != nil {
		c.breaker.RecordFailure()
		logger.From(ctx).Errorw("Failed to process message", "id", id, "err", err)
		_ = msg.Nack(false, true)
		return
	}

	c.breaker.RecordSuccess()

	if err := msg.Ack(false); err != nil {
		logger.From(ctx).Errorw("Failed to ack message", "err", err)
	}
}

func (c *Consumer) processIdempotently(ctx context.Context, messageID string, msg amqp.Delivery) error {
	processed, err := c.idempotency.IsProcessed(ctx, messageID)
	if err != nil {
		logger.From(ctx).Warnw("Idempotency check failed. Proceeding anyway.", "messageID", messageID, "err", err)
		processed = false
	}
	if processed {
		logger.From(ctx).Infow("Message already processed", "messageID", messageID)
		return nil
	}

	canProcess, err := c.idempotency.MarkAsProcessing(ctx, messageID)
	if err != nil {
		logger.From(ctx).Warnw("Mark as processing failed. Proceeding anyway.", "messageID", messageID, "err", err)
		canProcess = true // allow processing if store is unavailable
	}
	if !canProcess {
		logger.From(ctx).Infow("Message is being processed elsewhere", "messageID", messageID)
		return nil
	}

	defer func() {
		clearCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if clearErr := c.idempotency.ClearProcessing(clearCtx, messageID); clearErr != nil {
			logger.From(ctx).Warnw("Failed to clear processing lock", "err", clearErr)
		}
	}()

	var req domain.SendEmailRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	logger.From(ctx).Infow("Processing message", "messageID", messageID, "to", req.To, "template", req.Template)

	if err := c.useCase.Send(req); err != nil {
		return fmt.Errorf("use case handle failed: %w", err)
	}

	if err := c.idempotency.MarkAsProcessed(ctx, messageID); err != nil {
		logger.From(ctx).Warnw("Mark as processed failed. Proceeding anyway.", "messageID", messageID, "err", err)
	}

	logger.From(ctx).Infow("Message processed successfully", "messageID", messageID)
	return nil
}
