package weathergrpc

import (
	"context"
	"fmt"

	weatherpb2 "subscription/internal/adapter/weathergrpc/pb"
	"subscription/internal/domain"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"
)

type Client struct {
	conn   *grpc.ClientConn
	client weatherpb2.WeatherServiceClient
}

func NewClient(ctx context.Context, addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}
	return &Client{
		conn:   conn,
		client: weatherpb2.NewWeatherServiceClient(conn),
	}, nil
}

func (c *Client) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	resp, err := c.client.GetWeather(ctx, &weatherpb2.WeatherRequest{City: city})
	if err != nil {
		return domain.Report{}, err
	}
	return domain.Report{
		Temperature: resp.Temperature,
		Humidity:    int(resp.Humidity),
		Description: resp.Description,
	}, nil
}

func (c *Client) CityIsValid(ctx context.Context, city string) (bool, error) {
	resp, err := c.client.ValidateCity(ctx, &weatherpb2.ValidateRequest{City: city})
	if err != nil {
		return false, err
	}
	return resp.Valid, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
