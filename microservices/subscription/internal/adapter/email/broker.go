package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"subscription/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AsyncClient struct {
	channel  *amqp.Channel
	exchange string
	logger   *log.Logger
}

func NewAsyncClient(ch *amqp.Channel, exchange string, logger *log.Logger) *AsyncClient {
	return &AsyncClient{
		channel:  ch,
		exchange: exchange,
		logger:   logger,
	}
}

type Message struct {
	IdKey    string            `json:"id_key"`
	To       string            `json:"to"`
	Template string            `json:"template"`
	Data     map[string]string `json:"data"`
}

func (c *AsyncClient) SendConfirmationEmail(ctx context.Context, email, token, idKey string) error {
	msg := Message{
		IdKey:    idKey,
		To:       email,
		Template: "confirmation",
		Data: map[string]string{
			"token": token,
		},
	}
	return c.publish(ctx, "email.confirmation", msg)
}

func (c *AsyncClient) SendWeatherReport(ctx context.Context, email string, weather domain.Report, city, token, idKey string) error {
	msg := Message{
		IdKey:    idKey,
		To:       email,
		Template: "weather_report",
		Data: map[string]string{
			"temperature": fmt.Sprintf("%.2f", weather.Temperature),
			"humidity":    fmt.Sprintf("%d", weather.Humidity),
			"description": weather.Description,
			"city":        city,
			"token":       token,
		},
	}
	return c.publish(ctx, "email.weather_report", msg)
}

func (c *AsyncClient) publish(ctx context.Context, routingKey string, msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	pub := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		MessageId:    msg.IdKey,
		Timestamp:    time.Now(),
		DeliveryMode: amqp.Persistent,
	}

	if err := c.channel.PublishWithContext(
		ctx,
		c.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		pub,
	); err != nil {
		return fmt.Errorf("rabbitmq publish error: %w", err)
	}

	c.logger.Printf("Published to %s with idKey=%s", routingKey, msg.IdKey)
	return nil
}
