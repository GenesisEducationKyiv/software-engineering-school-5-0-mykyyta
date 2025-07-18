package infra

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitConn(url string) *amqp.Connection {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	return conn
}

func NewRabbitChannel(conn *amqp.Connection, exchange string) *amqp.Channel {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
	}

	err = ch.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}

	return ch
}
