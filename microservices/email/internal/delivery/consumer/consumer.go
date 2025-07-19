package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"email/internal/domain"

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

type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type EmailUseCase interface {
	Send(ctx context.Context, req domain.SendEmailRequest) error
}
type Consumer struct {
	source      MessageSource
	useCase     EmailUseCase
	idempotency IdempotencyStore
	logger      Logger
}

func NewConsumer(source MessageSource, useCase EmailUseCase, idempotency IdempotencyStore, logger Logger) *Consumer {
	return &Consumer{
		source:      source,
		useCase:     useCase,
		idempotency: idempotency,
		logger:      logger,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.source.Consume(ctx)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	c.logger.Println("Consumer started")

	for {
		select {
		case <-ctx.Done():
			c.logger.Println("Consumer stopped")
			return ctx.Err()

		case msg, ok := <-msgs:
			if !ok {
				c.logger.Println("Message channel closed")
				return nil
			}

			c.handle(ctx, msg)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, msg amqp.Delivery) {
	messageID := msg.MessageId
	if messageID == "" {
		c.logger.Printf("Skipping message without ID: %s", string(msg.Body))
		if err := msg.Nack(false, false); err != nil {
			c.logger.Printf("Failed to nack message: %v", err)
		}
		return
	}

	err := c.processIdempotently(ctx, messageID, msg)
	if err != nil {
		c.logger.Printf("Failed to process message [%s]: %v", messageID, err)
		if err := msg.Nack(false, true); err != nil {
			c.logger.Printf("Failed to nack message: %v", err)
		}
		return
	}

	if err := msg.Ack(false); err != nil {
		c.logger.Printf("Failed to ack message: %v", err)
	}
}

func (c *Consumer) processIdempotently(ctx context.Context, messageID string, msg amqp.Delivery) error {
	processed, err := c.idempotency.IsProcessed(ctx, messageID)
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if processed {
		c.logger.Printf("Message already processed: %s", messageID)
		return nil
	}

	canProcess, err := c.idempotency.MarkAsProcessing(ctx, messageID)
	if err != nil {
		return fmt.Errorf("mark as processing: %w", err)
	}
	if !canProcess {
		c.logger.Printf("Message %s is being processed elsewhere", messageID)
		return nil
	}

	defer func() {
		clearCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if clearErr := c.idempotency.ClearProcessing(clearCtx, messageID); clearErr != nil {
			c.logger.Printf("Failed to clear processing lock: %v", clearErr)
		}
	}()

	var req domain.SendEmailRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	c.logger.Printf("Processing message [%s] to %s (template: %s)", req.IdKey, req.To, req.Template)

	if err := c.useCase.Send(ctx, req); err != nil {
		return fmt.Errorf("use case handle failed: %w", err)
	}

	if err := c.idempotency.MarkAsProcessed(ctx, messageID); err != nil {
		return fmt.Errorf("mark as processed: %w", err)
	}

	c.logger.Printf("Message [%s] processed successfully", messageID)
	return nil
}
