package async

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	channel  *amqp.Channel
	exchange string
	logger   *log.Logger
}

func NewRabbitPublisher(ch *amqp.Channel, exchange string, logger *log.Logger) *RabbitPublisher {
	return &RabbitPublisher{channel: ch, exchange: exchange, logger: logger}
}

func (p *RabbitPublisher) Publish(ctx context.Context, routingKey string, msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	pub := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		Timestamp:    time.Now(),
		DeliveryMode: amqp.Persistent,
	}

	if withId, ok := msg.(interface{ GetIdKey() string }); ok {
		pub.MessageId = withId.GetIdKey()
	}

	if err := p.channel.PublishWithContext(ctx, p.exchange, routingKey, false, false, pub); err != nil {
		return fmt.Errorf("rabbitmq publish error: %w", err)
	}

	p.logger.Printf("Published to %s", routingKey)
	return nil
}
