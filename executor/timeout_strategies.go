package executor

import (
	"math"
	"time"
)

// FixedTimeoutStrategy implements a fixed timeout for all attempts
type FixedTimeoutStrategy struct {
	timeout time.Duration
}

// NewFixedTimeoutStrategy creates a new fixed timeout strategy
func NewFixedTimeoutStrategy(timeout time.Duration) *FixedTimeoutStrategy {
	return &FixedTimeoutStrategy{
		timeout: timeout,
	}
}

// GetTimeout returns the fixed timeout
func (f *FixedTimeoutStrategy) GetTimeout(attempt int) time.Duration {
	return f.timeout
}

// ProgressiveTimeoutStrategy implements increasing timeouts for retries
type ProgressiveTimeoutStrategy struct {
	baseTimeout time.Duration
	multiplier  float64
	maxTimeout  time.Duration
}

// NewProgressiveTimeoutStrategy creates a new progressive timeout strategy
func NewProgressiveTimeoutStrategy(baseTimeout time.Duration, multiplier float64, maxTimeout time.Duration) *ProgressiveTimeoutStrategy {
	return &ProgressiveTimeoutStrategy{
		baseTimeout: baseTimeout,
		multiplier:  multiplier,
		maxTimeout:  maxTimeout,
	}
}

// GetTimeout returns a progressively increasing timeout
func (p *ProgressiveTimeoutStrategy) GetTimeout(attempt int) time.Duration {
	timeout := time.Duration(float64(p.baseTimeout) * math.Pow(p.multiplier, float64(attempt-1)))

	if timeout > p.maxTimeout {
		timeout = p.maxTimeout
	}

	return timeout
}

// LinearTimeoutStrategy implements linearly increasing timeouts
type LinearTimeoutStrategy struct {
	baseTimeout time.Duration
	increment   time.Duration
	maxTimeout  time.Duration
}

// NewLinearTimeoutStrategy creates a new linear timeout strategy
func NewLinearTimeoutStrategy(baseTimeout, increment, maxTimeout time.Duration) *LinearTimeoutStrategy {
	return &LinearTimeoutStrategy{
		baseTimeout: baseTimeout,
		increment:   increment,
		maxTimeout:  maxTimeout,
	}
}

// GetTimeout returns a linearly increasing timeout
func (l *LinearTimeoutStrategy) GetTimeout(attempt int) time.Duration {
	timeout := l.baseTimeout + time.Duration(attempt-1)*l.increment

	if timeout > l.maxTimeout {
		timeout = l.maxTimeout
	}

	return timeout
}

// ConditionalTimeoutStrategy allows custom timeout logic
type ConditionalTimeoutStrategy struct {
	getTimeoutFn func(attempt int) time.Duration
}

// NewConditionalTimeoutStrategy creates a timeout strategy with custom logic
func NewConditionalTimeoutStrategy(getTimeoutFn func(attempt int) time.Duration) *ConditionalTimeoutStrategy {
	return &ConditionalTimeoutStrategy{
		getTimeoutFn: getTimeoutFn,
	}
}

// GetTimeout uses the custom timeout logic
func (c *ConditionalTimeoutStrategy) GetTimeout(attempt int) time.Duration {
	return c.getTimeoutFn(attempt)
}
