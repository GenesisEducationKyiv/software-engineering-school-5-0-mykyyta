package async

import (
	"context"
	"fmt"

	"subscription/internal/domain"
)

type Publisher interface {
	Publish(ctx context.Context, routingKey string, msg any) error
}

type Client struct {
	publisher Publisher
}

func NewAsyncClient(publisher Publisher) *Client {
	return &Client{publisher: publisher}
}

type EmailMessage struct {
	IdKey    string            `json:"-"`
	To       string            `json:"to"`
	Template string            `json:"template"`
	Data     map[string]string `json:"data"`
}

func (m EmailMessage) GetIdKey() string { return m.IdKey }

func (c *Client) SendConfirmationEmail(ctx context.Context, email, token, idKey string) error {
	msg := EmailMessage{
		IdKey:    idKey,
		To:       email,
		Template: "confirmation",
		Data:     map[string]string{"token": token},
	}
	return c.publisher.Publish(ctx, "email.confirmation", msg)
}

func (c *Client) SendWeatherReport(ctx context.Context, email string, weather domain.Report, city, token, idKey string) error {
	msg := EmailMessage{
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
