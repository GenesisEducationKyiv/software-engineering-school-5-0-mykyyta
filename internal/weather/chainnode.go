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

type ChainableProvider interface {
	provider
	SetNext(handler ChainableProvider) ChainableProvider
}

type ChainNode struct {
	next     ChainableProvider
	provider provider
}

func NewChainNode(p provider) *ChainNode {
	return &ChainNode{provider: p}
}

func (c *ChainNode) SetNext(next ChainableProvider) ChainableProvider {
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

	errAgg := err

	if c.next != nil {
		nextValid, errNext := c.next.CityIsValid(ctx, city)
		if errNext == nil {
			return nextValid, nil
		}
		errAgg = errors.Join(errAgg, errNext)
	}

	return false, errAgg
}
