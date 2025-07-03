package chain

import (
	"context"
	"errors"
	"fmt"
	"weatherApi/internal/domain"
)

type provider interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type ChainableProvider interface {
	provider
	SetNext(handler ChainableProvider) ChainableProvider
}

type Node struct {
	next     ChainableProvider
	provider provider
}

func NewNode(p provider) *Node {
	return &Node{provider: p}
}

func (c *Node) SetNext(next ChainableProvider) ChainableProvider {
	c.next = next
	return next
}

func (c *Node) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	report, err := c.provider.GetWeather(ctx, city)
	if err == nil {
		return report, nil
	}
	if c.next != nil {
		return c.next.GetWeather(ctx, city)
	}
	return domain.Report{}, fmt.Errorf("all providers failed: %w", err)
}

func (c *Node) CityIsValid(ctx context.Context, city string) (bool, error) {
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
