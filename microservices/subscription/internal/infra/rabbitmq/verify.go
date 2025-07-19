package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func VerifyExchange(ch *amqp.Channel, exchange, kind string) error {
	if err := ch.ExchangeDeclarePassive(
		exchange,
		kind, // "direct", "topic", тощо
		true, // durable
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("exchange %s does not exist or mismatched: %w", exchange, err)
	}
	return nil
}
