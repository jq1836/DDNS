package executor

import (
	"math"
	"time"
)

// ExponentialBackoffStrategy implements exponential backoff retry logic
type ExponentialBackoffStrategy struct {
	maxAttempts int
	baseDelay   time.Duration
	multiplier  float64
	maxDelay    time.Duration
}

// NewExponentialBackoffStrategy creates a new exponential backoff strategy
func NewExponentialBackoffStrategy(maxAttempts int, baseDelay time.Duration, multiplier float64) *ExponentialBackoffStrategy {
	return &ExponentialBackoffStrategy{
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
		multiplier:  multiplier,
		maxDelay:    30 * time.Second, // Default max delay
	}
}

// WithMaxDelay sets the maximum delay between retries
func (e *ExponentialBackoffStrategy) WithMaxDelay(maxDelay time.Duration) *ExponentialBackoffStrategy {
	e.maxDelay = maxDelay
	return e
}

// ShouldRetry determines if a task should be retried
func (e *ExponentialBackoffStrategy) ShouldRetry(attempt int, err error) bool {
	// Don't retry if we've reached max attempts
	if attempt >= e.maxAttempts {
		return false
	}

	// Retry on any error (this can be customized per use case)
	return err != nil
}

// GetDelay calculates the delay before the next retry using exponential backoff
func (e *ExponentialBackoffStrategy) GetDelay(attempt int) time.Duration {
	delay := time.Duration(float64(e.baseDelay) * math.Pow(e.multiplier, float64(attempt-1)))

	// Cap the delay at maxDelay
	if delay > e.maxDelay {
		delay = e.maxDelay
	}

	return delay
}

// GetMaxAttempts returns the maximum number of attempts
func (e *ExponentialBackoffStrategy) GetMaxAttempts() int {
	return e.maxAttempts
}

// LinearBackoffStrategy implements linear backoff retry logic
type LinearBackoffStrategy struct {
	maxAttempts int
	baseDelay   time.Duration
	increment   time.Duration
}

// NewLinearBackoffStrategy creates a new linear backoff strategy
func NewLinearBackoffStrategy(maxAttempts int, baseDelay, increment time.Duration) *LinearBackoffStrategy {
	return &LinearBackoffStrategy{
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
		increment:   increment,
	}
}

// ShouldRetry determines if a task should be retried
func (l *LinearBackoffStrategy) ShouldRetry(attempt int, err error) bool {
	if attempt >= l.maxAttempts {
		return false
	}
	return err != nil
}

// GetDelay calculates the delay before the next retry using linear backoff
func (l *LinearBackoffStrategy) GetDelay(attempt int) time.Duration {
	return l.baseDelay + time.Duration(attempt-1)*l.increment
}

// GetMaxAttempts returns the maximum number of attempts
func (l *LinearBackoffStrategy) GetMaxAttempts() int {
	return l.maxAttempts
}

// FixedDelayStrategy implements fixed delay retry logic
type FixedDelayStrategy struct {
	maxAttempts int
	delay       time.Duration
}

// NewFixedDelayStrategy creates a new fixed delay strategy
func NewFixedDelayStrategy(maxAttempts int, delay time.Duration) *FixedDelayStrategy {
	return &FixedDelayStrategy{
		maxAttempts: maxAttempts,
		delay:       delay,
	}
}

// ShouldRetry determines if a task should be retried
func (f *FixedDelayStrategy) ShouldRetry(attempt int, err error) bool {
	if attempt >= f.maxAttempts {
		return false
	}
	return err != nil
}

// GetDelay returns the fixed delay
func (f *FixedDelayStrategy) GetDelay(attempt int) time.Duration {
	return f.delay
}

// GetMaxAttempts returns the maximum number of attempts
func (f *FixedDelayStrategy) GetMaxAttempts() int {
	return f.maxAttempts
}

// NoRetryStrategy implements no retry logic (fail fast)
type NoRetryStrategy struct{}

// NewNoRetryStrategy creates a strategy that never retries
func NewNoRetryStrategy() *NoRetryStrategy {
	return &NoRetryStrategy{}
}

// ShouldRetry always returns false
func (n *NoRetryStrategy) ShouldRetry(attempt int, err error) bool {
	return false
}

// GetDelay returns zero delay (not used since we don't retry)
func (n *NoRetryStrategy) GetDelay(attempt int) time.Duration {
	return 0
}

// GetMaxAttempts returns 1 (only the initial attempt)
func (n *NoRetryStrategy) GetMaxAttempts() int {
	return 1
}

// ConditionalRetryStrategy allows custom retry conditions
type ConditionalRetryStrategy struct {
	maxAttempts   int
	baseDelay     time.Duration
	shouldRetryFn func(attempt int, err error) bool
	getDelayFn    func(attempt int) time.Duration
}

// NewConditionalRetryStrategy creates a retry strategy with custom logic
func NewConditionalRetryStrategy(
	maxAttempts int,
	baseDelay time.Duration,
	shouldRetryFn func(attempt int, err error) bool,
	getDelayFn func(attempt int) time.Duration,
) *ConditionalRetryStrategy {
	return &ConditionalRetryStrategy{
		maxAttempts:   maxAttempts,
		baseDelay:     baseDelay,
		shouldRetryFn: shouldRetryFn,
		getDelayFn:    getDelayFn,
	}
}

// ShouldRetry uses the custom retry logic
func (c *ConditionalRetryStrategy) ShouldRetry(attempt int, err error) bool {
	if attempt >= c.maxAttempts {
		return false
	}
	if c.shouldRetryFn != nil {
		return c.shouldRetryFn(attempt, err)
	}
	return err != nil
}

// GetDelay uses custom delay logic or falls back to base delay
func (c *ConditionalRetryStrategy) GetDelay(attempt int) time.Duration {
	if c.getDelayFn != nil {
		return c.getDelayFn(attempt)
	}
	return c.baseDelay
}

// GetMaxAttempts returns the maximum number of attempts
func (c *ConditionalRetryStrategy) GetMaxAttempts() int {
	return c.maxAttempts
}
