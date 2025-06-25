package weather

import (
	"context"
	"errors"
	"fmt"
)

type provider interface {
	GetWeather(ctx context.Context, city string) (Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type ChainableHandler interface {
	SetNext(handler ChainableHandler) ChainableHandler
	GetWeather(ctx context.Context, city string) (Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type ChainNode struct {
	next     ChainableHandler
	provider provider
}

func NewChainNode(p provider) *ChainNode {
	return &ChainNode{provider: p}
}

func (c *ChainNode) SetNext(next ChainableHandler) ChainableHandler {
	c.next = next
	return next
}

func (c *ChainNode) GetWeather(ctx context.Context, city string) (Report, error) {
	report, err := c.provider.GetWeather(ctx, city)
	if err == nil {
		return report, nil
	}
	if c.next != nil {
		return c.next.GetWeather(ctx, city)
	}
	return Report{}, fmt.Errorf("all providers failed: %w", err)
}

func (c *ChainNode) CityIsValid(ctx context.Context, city string) (bool, error) {
	valid, err := c.provider.CityIsValid(ctx, city)
	if err == nil {
		return valid, nil
	}

	if c.next == nil {
		return false, err
	}

	nextValid, nextErr := c.next.CityIsValid(ctx, city)
	if nextErr == nil {
		return nextValid, nil
	}

	// Priority: return ErrCityNotFound if any provider reports city not found
	if errors.Is(err, ErrCityNotFound) || errors.Is(nextErr, ErrCityNotFound) {
		return false, ErrCityNotFound
	}

	return false, err
}
