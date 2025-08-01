package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageSource struct {
	channel   *amqp.Channel
	queueName string
}

func NewSource(conn *Connection, queueName string) *MessageSource {
	return &MessageSource{
		channel:   conn.Channel(),
		queueName: queueName,
	}
}

func (r *MessageSource) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	msgs, err := r.channel.Consume(
		r.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming from queue %s: %w", r.queueName, err)
	}
	return msgs, nil
}
