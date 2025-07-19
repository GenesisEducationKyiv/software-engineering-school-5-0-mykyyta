package di

import (
	"context"
	"log"
	"time"

	"email/internal/adapter/idempotency"
	"email/internal/config"
	"email/internal/delivery/consumer"
	"email/internal/email"
	"email/internal/infra/rabbitmq"

	"github.com/redis/go-redis/v9"
)

type QueueModule struct {
	Consumer     *consumer.Consumer
	RabbitConn   *rabbitmq.Connection
	ShutdownFunc func() error
}

func NewQueueModule(ctx context.Context, cfg *config.Config, svc email.Service, redisClient *redis.Client, logger *log.Logger) (*QueueModule, error) {
	rmqConn, err := rabbitmq.NewConnection("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}

	err = rabbitmq.Setup(
		rmqConn.Channel(),
		"email_exchange",
		[]rabbitmq.QueueConfig{
			{
				QueueName:   "email_queue",
				RoutingKeys: []string{"email.confirmation", "email.weather_report"},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	source := rabbitmq.NewSource(rmqConn, "email_queue")
	store := idempotency.NewRedisStore(redisClient, 24*time.Hour)
	cons := consumer.NewConsumer(source, svc, store, logger)

	return &QueueModule{
		Consumer:   cons,
		RabbitConn: rmqConn,
		ShutdownFunc: func() error {
			return rmqConn.Close()
		},
	}, nil
}
