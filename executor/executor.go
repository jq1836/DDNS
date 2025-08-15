package executor

import (
	"context"
	"time"
)

// Task represents a generic operation that can be executed with retry and timeout logic
type Task[T any] func(ctx context.Context) (T, error)

// Result represents the result of a task execution
type Result[T any] struct {
	Value   T
	Error   error
	Attempt int
}

// RetryStrategy defines the interface for retry strategies
type RetryStrategy interface {
	// ShouldRetry determines if a task should be retried based on the attempt number and error
	ShouldRetry(attempt int, err error) bool
	// GetDelay returns the delay before the next retry attempt
	GetDelay(attempt int) time.Duration
	// GetMaxAttempts returns the maximum number of attempts (including the initial attempt)
	GetMaxAttempts() int
}

// TimeoutStrategy defines the interface for timeout strategies
type TimeoutStrategy interface {
	// GetTimeout returns the timeout for a task based on the attempt number
	GetTimeout(attempt int) time.Duration
}

// Executor executes tasks with retry and timeout strategies
type Executor struct {
	retryStrategy   RetryStrategy
	timeoutStrategy TimeoutStrategy
	onRetry         func(attempt int, err error, delay time.Duration) // Optional callback for retry events
	onTimeout       func(attempt int, timeout time.Duration)          // Optional callback for timeout events
}

// ExecutorOption defines a function type for configuring the executor
type ExecutorOption func(*Executor)

// NewExecutor creates a new task executor with strategies
func NewExecutor(options ...ExecutorOption) *Executor {
	executor := &Executor{
		retryStrategy:   NewExponentialBackoffStrategy(3, time.Second, 2.0),
		timeoutStrategy: NewFixedTimeoutStrategy(30 * time.Second),
	}

	for _, option := range options {
		option(executor)
	}

	return executor
}

// WithRetryStrategy sets the retry strategy
func WithRetryStrategy(strategy RetryStrategy) ExecutorOption {
	return func(e *Executor) {
		e.retryStrategy = strategy
	}
}

// WithTimeoutStrategy sets the timeout strategy
func WithTimeoutStrategy(strategy TimeoutStrategy) ExecutorOption {
	return func(e *Executor) {
		e.timeoutStrategy = strategy
	}
}

// WithRetryCallback sets a callback that's called before each retry
func WithRetryCallback(callback func(attempt int, err error, delay time.Duration)) ExecutorOption {
	return func(e *Executor) {
		e.onRetry = callback
	}
}

// WithTimeoutCallback sets a callback that's called when a timeout occurs
func WithTimeoutCallback(callback func(attempt int, timeout time.Duration)) ExecutorOption {
	return func(e *Executor) {
		e.onTimeout = callback
	}
}

// Execute executes a task with retry and timeout logic
func Execute[T any](executor *Executor, ctx context.Context, task Task[T]) (*Result[T], error) {
	var lastResult Result[T]
	maxAttempts := executor.retryStrategy.GetMaxAttempts()

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Create a context with timeout for this attempt
		timeout := executor.timeoutStrategy.GetTimeout(attempt)
		taskCtx, cancel := context.WithTimeout(ctx, timeout)

		// Notify about timeout if callback is set
		if executor.onTimeout != nil {
			executor.onTimeout(attempt, timeout)
		}

		// Execute the task
		value, err := task(taskCtx)
		cancel() // Clean up the context

		lastResult = Result[T]{
			Value:   value,
			Error:   err,
			Attempt: attempt,
		}

		// If successful, return immediately
		if err == nil {
			return &lastResult, nil
		}

		// Check if we should retry
		if !executor.retryStrategy.ShouldRetry(attempt, err) {
			break
		}

		// If this isn't the last attempt, wait before retrying
		if attempt < maxAttempts {
			delay := executor.retryStrategy.GetDelay(attempt)

			// Notify about retry if callback is set
			if executor.onRetry != nil {
				executor.onRetry(attempt, err, delay)
			}

			// Wait with context cancellation support
			select {
			case <-ctx.Done():
				lastResult.Error = ctx.Err()
				return &lastResult, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	// Return the last result
	return &lastResult, lastResult.Error
}

// ExecuteSimple is a convenience function that returns just the value and error
func ExecuteSimple[T any](executor *Executor, ctx context.Context, task Task[T]) (T, error) {
	result, err := Execute(executor, ctx, task)
	if err != nil {
		var zero T
		return zero, err
	}
	return result.Value, result.Error
}

// ExecuteWithTimeout executes a task with a simple timeout (no retries)
func ExecuteWithTimeout[T any](ctx context.Context, timeout time.Duration, task Task[T]) (T, error) {
	executor := NewExecutor(
		WithRetryStrategy(NewNoRetryStrategy()),
		WithTimeoutStrategy(NewFixedTimeoutStrategy(timeout)),
	)
	return ExecuteSimple(executor, ctx, task)
}

// ExecuteWithRetries executes a task with retries but a fixed timeout
func ExecuteWithRetries[T any](ctx context.Context, maxRetries int, delay time.Duration, timeout time.Duration, task Task[T]) (T, error) {
	executor := NewExecutor(
		WithRetryStrategy(NewFixedDelayStrategy(maxRetries+1, delay)),
		WithTimeoutStrategy(NewFixedTimeoutStrategy(timeout)),
	)
	return ExecuteSimple(executor, ctx, task)
}
