package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueConfig struct {
	QueueName   string
	RoutingKeys []string
}

func Setup(channel *amqp.Channel, exchangeName string, queues []QueueConfig) error {
	err := channel.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	for _, q := range queues {
		_, err := channel.QueueDeclare(
			q.QueueName,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("declare queue %s: %w", q.QueueName, err)
		}

		for _, key := range q.RoutingKeys {
			err := channel.QueueBind(
				q.QueueName,
				key,
				exchangeName,
				false,
				nil,
			)
			if err != nil {
				return fmt.Errorf("bind queue %s with key %s: %w", q.QueueName, key, err)
			}
		}
	}

	return nil
}
