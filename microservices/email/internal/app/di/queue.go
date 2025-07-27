package di

import (
	"context"
	"time"

	"email/internal/adapter/idempotency"
	"email/internal/config"
	"email/internal/delivery/consumer"
	"email/internal/email"
	"email/internal/infra/rabbitmq"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type QueueModule struct {
	Consumer     *consumer.Consumer
	RabbitConn   *rabbitmq.Connection
	ShutdownFunc func() error
}

func NewQueueModule(ctx context.Context, cfg *config.Config, svc email.Service, redisClient *redis.Client) (*QueueModule, error) {
	rmqConn, err := rabbitmq.NewConnection(cfg.RabbitMQURL)
	if err != nil {
		logger.From(ctx).Errorw("Failed to connect to RabbitMQ", "err", err)
		return nil, err
	}

	err = rabbitmq.Setup(
		rmqConn.Channel(),
		"email.exchange",
		[]rabbitmq.QueueConfig{
			{
				QueueName:   "email.queue",
				RoutingKeys: []string{"email.confirmation", "email.weather_report"},
			},
		},
	)
	if err != nil {
		logger.From(ctx).Errorw("Failed to setup RabbitMQ", "err", err)
		err = rmqConn.Close()
		if err != nil {
			logger.From(ctx).Errorw("Error during RabbitMQ connection close", "err", err)
		}
		return nil, err
	}

	source := rabbitmq.NewSource(rmqConn, "email.queue")
	store := idempotency.NewRedisStore(redisClient, 24*time.Hour)
	breaker := consumer.NewDefaultCB()
	cons := consumer.NewConsumer(source, svc, store, breaker)

	return &QueueModule{
		Consumer:   cons,
		RabbitConn: rmqConn,
		ShutdownFunc: func() error {
			if err := rmqConn.Close(); err != nil {
				logger.From(ctx).Errorw("Error during RabbitMQ connection close", "err", err)
				return err
			}
			return nil
		},
	}, nil
}
