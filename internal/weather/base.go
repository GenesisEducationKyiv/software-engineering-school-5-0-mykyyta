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

type BaseProvider interface {
	SetNext(handler BaseProvider) BaseProvider
	GetWeather(ctx context.Context, city string) (Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type Base struct {
	next     BaseProvider
	provider provider
}

func NewBase(p provider) *Base {
	return &Base{provider: p}
}

func (h *Base) SetNext(next BaseProvider) BaseProvider {
	h.next = next
	return next
}

func (h *Base) GetWeather(ctx context.Context, city string) (Report, error) {
	data, err := h.provider.GetWeather(ctx, city)
	if err == nil {
		return data, nil
	}
	if h.next != nil {
		return h.next.GetWeather(ctx, city)
	}
	return Report{}, fmt.Errorf("all providers failed: %w", err)
}

func (h *Base) CityIsValid(ctx context.Context, city string) (bool, error) {
	var lastErr error
	var cityNotFoundSeen bool

	current := h
	for current != nil {
		ok, err := current.provider.CityIsValid(ctx, city)
		if err == nil {
			return ok, nil
		}
		if errors.Is(err, ErrCityNotFound) {
			cityNotFoundSeen = true
		} else {
			lastErr = err
		}

		if next, ok := current.next.(*Base); ok {
			current = next
		} else {
			current = nil
		}
	}

	if cityNotFoundSeen {
		return false, ErrCityNotFound
	}
	return false, fmt.Errorf("all providers failed: %w", lastErr)
}
