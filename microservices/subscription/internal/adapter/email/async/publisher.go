package async

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	loggerPkg "subscription/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	channel  *amqp.Channel
	exchange string
}

type IdKeyGetter interface {
	GetIdKey() string
}

func NewRabbitPublisher(ch *amqp.Channel, exchange string) *RabbitPublisher {
	return &RabbitPublisher{channel: ch, exchange: exchange}
}

func (p *RabbitPublisher) Publish(ctx context.Context, routingKey string, msg IdKeyGetter) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	pub := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		Timestamp:    time.Now(),
		DeliveryMode: amqp.Persistent,
		MessageId:    msg.GetIdKey(),
	}

	if err := p.channel.PublishWithContext(ctx, p.exchange, routingKey, false, false, pub); err != nil {
		return fmt.Errorf("rabbitmq publish error: %w", err)
	}

	logger := loggerPkg.From(ctx)
	logger.Infow("Published message", "routingKey", routingKey, "msgId", msg.GetIdKey())
	return nil
}
