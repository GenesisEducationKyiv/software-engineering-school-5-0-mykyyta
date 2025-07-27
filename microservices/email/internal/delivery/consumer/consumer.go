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
	CanExecute(ctx context.Context) bool
	RecordSuccess(ctx context.Context)
	RecordFailure(ctx context.Context)
}

type EmailUseCase interface {
	Send(ctx context.Context, req domain.SendEmailRequest) error
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
		logger.From(ctx).Warnw("Skipping message without ID", "routing_key", msg.RoutingKey)
		_ = msg.Nack(false, false)
		return
	}

	// Додаємо message_id до логера в контексті
	log := logger.From(ctx).With("message_id", id)
	ctx = logger.With(ctx, log)

	logger.From(ctx).Infow("Message received", "routing_key", msg.RoutingKey, "body_size", len(msg.Body))

	if !c.breaker.CanExecute(ctx) {
		logger.From(ctx).Warnw("Circuit breaker is open, rejecting message")
		delay := time.Duration(500+rand.Intn(500)) * time.Millisecond
		time.Sleep(delay)
		_ = msg.Nack(false, true)
		return
	}

	start := time.Now()
	err := c.processIdempotently(ctx, id, msg)
	duration := time.Since(start)

	if err != nil {
		c.breaker.RecordFailure(ctx)
		logger.From(ctx).Errorw("Message processing failed", "err", err, "duration_ms", duration.Milliseconds())
		_ = msg.Nack(false, true)
		return
	}

	c.breaker.RecordSuccess(ctx)
	logger.From(ctx).Infow("Message processed successfully", "duration_ms", duration.Milliseconds())

	if err := msg.Ack(false); err != nil {
		logger.From(ctx).Errorw("Failed to ack message", "err", err)
	}
}

func (c *Consumer) processIdempotently(ctx context.Context, messageID string, msg amqp.Delivery) error {
	processed, err := c.idempotency.IsProcessed(ctx, messageID)
	if err != nil {
		logger.From(ctx).Errorw("Idempotency check failed, proceeding with processing", "err", err)
		processed = false
	}
	if processed {
		logger.From(ctx).Infow("Message already processed, skipping")
		return nil
	}

	canProcess, err := c.idempotency.MarkAsProcessing(ctx, messageID)
	if err != nil {
		logger.From(ctx).Errorw("Failed to mark message as processing, proceeding anyway", "err", err)
		canProcess = true // allow processing if store is unavailable
	}
	if !canProcess {
		logger.From(ctx).Infow("Message is being processed by another worker, skipping")
		return nil
	}

	defer func() {
		clearCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if clearErr := c.idempotency.ClearProcessing(clearCtx, messageID); clearErr != nil {
			logger.From(ctx).Errorw("Failed to clear processing lock", "err", clearErr)
		}
	}()

	var req domain.SendEmailRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	logger.From(ctx).Infow("Starting email processing", "to", req.To, "template", req.Template)

	if err := c.useCase.Send(ctx, req); err != nil {
		return fmt.Errorf("email sending failed: %w", err)
	}

	if err := c.idempotency.MarkAsProcessed(ctx, messageID); err != nil {
		logger.From(ctx).Errorw("Failed to mark message as processed, email was sent successfully", "err", err)
	}

	logger.From(ctx).Infow("Email sent successfully", "to", req.To, "template", req.Template)
	return nil
}
