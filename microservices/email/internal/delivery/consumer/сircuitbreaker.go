package consumer

import (
	"context"
	"sync"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type simpleCB struct {
	mu           sync.Mutex
	failures     int
	maxFailures  int
	openUntil    time.Time
	openDuration time.Duration
}

func NewCB(maxFailures int, openDuration time.Duration) CircuitBreaker {
	return &simpleCB{
		maxFailures:  maxFailures,
		openDuration: openDuration,
	}
}

func NewDefaultCB() CircuitBreaker {
	return NewCB(2, 5*time.Second)
}

func (c *simpleCB) CanExecute(ctx context.Context) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.openUntil) {
		return false
	}

	if !c.openUntil.IsZero() {
		loggerPkg.From(ctx).Info("Circuit breaker recovered, allowing execution")
		c.reset()
	}
	return true
}

func (c *simpleCB) RecordSuccess(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	wasOpen := !c.openUntil.IsZero()
	c.reset()

	if wasOpen {
		loggerPkg.From(ctx).Info("Circuit breaker closed after successful execution")
	}
}

func (c *simpleCB) RecordFailure(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures++
	loggerPkg.From(ctx).Warn("Circuit breaker recorded failure", "failures", c.failures, "max_failures", c.maxFailures)

	if c.failures >= c.maxFailures {
		c.openUntil = time.Now().Add(c.openDuration)
		loggerPkg.From(ctx).Error("Circuit breaker opened due to too many failures",
			"failures", c.failures,
			"max_failures", c.maxFailures,
			"open_until", c.openUntil,
			"open_duration_seconds", c.openDuration.Seconds())
	}
}

func (c *simpleCB) reset() {
	c.failures = 0
	c.openUntil = time.Time{}
}
