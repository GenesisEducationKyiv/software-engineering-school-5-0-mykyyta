package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"email/internal/domain"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"github.com/google/uuid"
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

	logger := loggerPkg.From(ctx)
	logger.Info("Consumer started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Consumer stopped")
			return ctx.Err()

		case msg, ok := <-msgs:
			if !ok {
				logger.Info("Message channel closed")
				return nil
			}

			c.handle(ctx, msg)
		}
	}
}

func extractCorrelationID(msgBody []byte) string {
	var msgData struct {
		CorrelationID string `json:"correlation_id"`
	}

	if err := json.Unmarshal(msgBody, &msgData); err == nil && msgData.CorrelationID != "" {
		return msgData.CorrelationID
	}

	// Якщо немає correlation ID, генеруємо новий для цього повідомлення
	return "email-" + uuid.New().String()[:8]
}

func (c *Consumer) handle(ctx context.Context, msg amqp.Delivery) {
	logger := loggerPkg.From(ctx)

	messageId := msg.MessageId
	if messageId == "" {
		logger.Warn("Skipping message without ID", "routing_key", msg.RoutingKey)
		_ = msg.Nack(false, false)
		return
	}

	correlationID := extractCorrelationID(msg.Body)

	enrichedLogger := logger.With(
		"message_id", messageId,
		"correlation_id", correlationID,
	)

	ctx = loggerPkg.WithCorrelationID(ctx, correlationID)
	ctx = loggerPkg.With(ctx, enrichedLogger)

	enrichedLogger.Debug("Message received", "routing_key", msg.RoutingKey)

	if !c.breaker.CanExecute(ctx) {
		enrichedLogger.Warn("Circuit breaker is open, rejecting message")
		delay := time.Duration(500+rand.Intn(500)) * time.Millisecond
		time.Sleep(delay)
		_ = msg.Nack(false, true)
		return
	}

	start := time.Now()
	err := c.processIdempotently(ctx, messageId, msg)
	duration := time.Since(start)

	if err != nil {
		c.breaker.RecordFailure(ctx)
		enrichedLogger.Error("Message processing failed",
			"error_chain", err.Error(),
			"routing_key", msg.RoutingKey,
			"duration_ms", duration.Milliseconds())
		_ = msg.Nack(false, true)
		return
	}

	c.breaker.RecordSuccess(ctx)
	if duration > 5*time.Second {
		enrichedLogger.Warn("Slow message processing", "duration_ms", duration.Milliseconds())
	}

	if err := msg.Ack(false); err != nil {
		enrichedLogger.Error("Failed to ack message", "err", err)
	}
}

func (c *Consumer) processIdempotently(ctx context.Context, messageID string, msg amqp.Delivery) error {
	logger := loggerPkg.From(ctx)

	processed, err := c.idempotency.IsProcessed(ctx, messageID)
	if err != nil {
		logger.Error("Idempotency check failed, proceeding with processing", "err", err)
		processed = false
	}
	if processed {
		logger.Debug("Message already processed, skipping")
		return nil
	}

	canProcess, err := c.idempotency.MarkAsProcessing(ctx, messageID)
	if err != nil {
		logger.Error("Failed to mark message as processing, proceeding anyway", "err", err)
		canProcess = true
	}
	if !canProcess {
		logger.Debug("Message is being processed by another worker, skipping")
		return nil
	}

	defer func() {
		clearCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if clearErr := c.idempotency.ClearProcessing(clearCtx, messageID); clearErr != nil {
			logger.Error("Failed to clear processing lock", "err", clearErr)
		}
	}()

	var req domain.SendEmailRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	if err := c.useCase.Send(ctx, req); err != nil {
		return fmt.Errorf("email sending failed: %w", err)
	}

	if err := c.idempotency.MarkAsProcessed(ctx, messageID); err != nil {
		logger.Error("Failed to mark message as processed, email was sent successfully", "err", err)
	}

	return nil
}
