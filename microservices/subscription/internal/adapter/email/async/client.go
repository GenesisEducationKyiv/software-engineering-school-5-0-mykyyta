package async

import (
	"context"
	"fmt"

	"subscription/internal/domain"
)

type Publisher interface {
	Publish(ctx context.Context, routingKey string, msg any) error
}

type AsyncClient struct {
	publisher Publisher
}

func NewAsyncClient(publisher Publisher) *AsyncClient {
	return &AsyncClient{publisher: publisher}
}

type Message struct {
	IdKey    string            `json:"idKey"`
	To       string            `json:"to"`
	Template string            `json:"template"`
	Data     map[string]string `json:"data"`
}

func (m Message) GetIdKey() string { return m.IdKey }

func (c *AsyncClient) SendConfirmationEmail(ctx context.Context, email, token, idKey string) error {
	msg := Message{
		IdKey:    idKey,
		To:       email,
		Template: "confirmation",
		Data:     map[string]string{"token": token},
	}
	return c.publisher.Publish(ctx, "email.confirmation", msg)
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
	return c.publisher.Publish(ctx, "email.weather_report", msg)
}
