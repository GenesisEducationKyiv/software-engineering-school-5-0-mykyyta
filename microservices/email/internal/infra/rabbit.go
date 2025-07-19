package infra

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConnection(url string) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Connection{conn: conn, channel: ch}, nil
}

func (c *Connection) Close() error {
	_ = c.channel.Close()
	return c.conn.Close()
}

func (c *Connection) Channel() *amqp.Channel {
	return c.channel
}

// RabbitMQSource реалізує MessageSource
type RabbitMQSource struct {
	channel   *amqp.Channel
	queueName string
}

func NewRabbitMQSource(conn *Connection, queueName string) *RabbitMQSource {
	return &RabbitMQSource{
		channel:   conn.Channel(),
		queueName: queueName,
	}
}

func (r *RabbitMQSource) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		r.queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}
