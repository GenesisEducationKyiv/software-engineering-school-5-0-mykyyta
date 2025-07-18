package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"email/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	IdKey    string            `json:"idKey"`
	To       string            `json:"to"`
	Template string            `json:"template"`
	Data     map[string]string `json:"data"`
}

type EmailSender interface {
	Send(req domain.SendEmailRequest) error
}

type IdempotencyStore interface {
	IsProcessed(ctx context.Context, messageID string) (bool, error)
	MarkAsProcessed(ctx context.Context, messageID string) error
	MarkAsProcessing(ctx context.Context, messageID string) (bool, error)
	ClearProcessing(ctx context.Context, messageID string) error
}

// Інтерфейс для тестування.
type MessageSource interface {
	Consume(ctx context.Context) (<-chan amqp.Delivery, error)
}

type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type Consumer struct {
	messageSource MessageSource
	emailSender   EmailSender
	idempotency   IdempotencyStore
	logger        Logger
}

func NewConsumer(
	messageSource MessageSource,
	emailSender EmailSender,
	idempotency IdempotencyStore,
	logger Logger,
) *Consumer {
	if logger == nil {
		logger = log.New(os.Stdout, "[CONSUMER] ", log.LstdFlags)
	}

	return &Consumer{
		messageSource: messageSource,
		emailSender:   emailSender,
		idempotency:   idempotency,
		logger:        logger,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	messages, err := c.messageSource.Consume(ctx)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	c.logger.Println("Consumer started")

	for {
		select {
		case <-ctx.Done():
			c.logger.Println("Consumer stopped")
			return ctx.Err()
		case msg, ok := <-messages:
			if !ok {
				c.logger.Println("Message channel closed")
				return nil
			}
			c.handleMessage(ctx, msg)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	messageID := msg.MessageId
	if messageID == "" {
		c.logger.Printf("Message without ID, skipping: %s", string(msg.Body))
		msg.Nack(false, false)
		return
	}

	if err := c.processIdempotently(ctx, messageID, msg); err != nil {
		c.logger.Printf("Processing failed for %s: %v", messageID, err)
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
}

func (c *Consumer) processIdempotently(ctx context.Context, messageID string, msg amqp.Delivery) error {
	processed, err := c.idempotency.IsProcessed(ctx, messageID)
	if err != nil {
		return fmt.Errorf("check processed: %w", err)
	}

	if processed {
		c.logger.Printf("Message %s already processed, skipping", messageID)
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
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if clearErr := c.idempotency.ClearProcessing(cleanupCtx, messageID); clearErr != nil {
			c.logger.Printf("Failed to clear processing lock: %v", clearErr)
		}
	}()

	}()

	if err := c.processMessage(msg); err != nil {
		return fmt.Errorf("process message: %w", err)
	}

	if err := c.idempotency.MarkAsProcessed(ctx, messageID); err != nil {
		return fmt.Errorf("mark as processed: %w", err)
	}

	return nil
}

func (c *Consumer) processMessage(msg amqp.Delivery) error {
	var emailMsg Message
	if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	c.logger.Printf("Processing email [%s] to %s", emailMsg.IdKey, emailMsg.To)

	req := domain.SendEmailRequest{
		IdKey:    emailMsg.IdKey,
		To:       emailMsg.To,
		Template: domain.TemplateName(emailMsg.Template),
		Data:     emailMsg.Data,
	}

	if err := c.emailSender.Send(req); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	c.logger.Printf("Email sent successfully [%s]", emailMsg.IdKey)
	return nil
}
