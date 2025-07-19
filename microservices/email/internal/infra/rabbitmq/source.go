package rabbitmq

import (
	"context"

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
	return r.channel.Consume(
		r.queueName, // queue
		"",          // consumer name
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
}
