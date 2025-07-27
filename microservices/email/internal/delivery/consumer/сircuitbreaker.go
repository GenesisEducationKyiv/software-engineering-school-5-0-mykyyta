package consumer

import (
	"sync"
	"time"
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

func (c *simpleCB) CanExecute() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.openUntil) {
		return false
	}

	if !c.openUntil.IsZero() {
		c.reset()
	}
	return true
}

func (c *simpleCB) RecordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reset()
}

func (c *simpleCB) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures++
	if c.failures >= c.maxFailures {
		c.openUntil = time.Now().Add(c.openDuration)
	}
}

func (c *simpleCB) reset() {
	c.failures = 0
	c.openUntil = time.Time{}
}
