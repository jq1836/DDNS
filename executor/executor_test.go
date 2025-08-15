package executor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecutorWithSuccessfulTask(t *testing.T) {
	executor := NewExecutor()

	task := func(ctx context.Context) (string, error) {
		return "success", nil
	}

	result, err := Execute(executor, context.Background(), task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Value != "success" {
		t.Errorf("Expected 'success', got '%s'", result.Value)
	}

	if result.Attempt != 1 {
		t.Errorf("Expected 1 attempt, got %d", result.Attempt)
	}
}

func TestExecutorWithRetries(t *testing.T) {
	attempts := 0
	task := func(ctx context.Context) (int, error) {
		attempts++
		if attempts < 3 {
			return 0, errors.New("temporary failure")
		}
		return 42, nil
	}

	executor := NewExecutor(
		WithRetryStrategy(NewFixedDelayStrategy(3, 10*time.Millisecond)),
	)

	result, err := Execute(executor, context.Background(), task)
	if err != nil {
		t.Fatalf("Expected no error after retries, got %v", err)
	}

	if result.Value != 42 {
		t.Errorf("Expected 42, got %d", result.Value)
	}

	if result.Attempt != 3 {
		t.Errorf("Expected 3 attempts, got %d", result.Attempt)
	}
}

func TestExecutorMaxRetriesExceeded(t *testing.T) {
	task := func(ctx context.Context) (string, error) {
		return "", errors.New("persistent failure")
	}

	executor := NewExecutor(
		WithRetryStrategy(NewFixedDelayStrategy(2, 10*time.Millisecond)),
	)

	result, err := Execute(executor, context.Background(), task)
	if err == nil {
		t.Fatal("Expected error when max retries exceeded")
	}

	if result.Attempt != 2 {
		t.Errorf("Expected 2 attempts, got %d", result.Attempt)
	}
}

func TestExecutorWithTimeout(t *testing.T) {
	task := func(ctx context.Context) (string, error) {
		// Simulate a long-running task
		select {
		case <-time.After(100 * time.Millisecond):
			return "completed", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	executor := NewExecutor(
		WithTimeoutStrategy(NewFixedTimeoutStrategy(50*time.Millisecond)),
		WithRetryStrategy(NewNoRetryStrategy()),
	)

	result, err := Execute(executor, context.Background(), task)
	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if !errors.Is(result.Error, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", result.Error)
	}
}

func TestExecuteWithTimeout(t *testing.T) {
	task := func(ctx context.Context) (string, error) {
		return "fast", nil
	}

	result, err := ExecuteWithTimeout(context.Background(), time.Second, task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "fast" {
		t.Errorf("Expected 'fast', got '%s'", result)
	}
}

func TestExecuteWithRetries(t *testing.T) {
	attempts := 0
	task := func(ctx context.Context) (bool, error) {
		attempts++
		if attempts < 2 {
			return false, errors.New("retry me")
		}
		return true, nil
	}

	result, err := ExecuteWithRetries(context.Background(), 2, 10*time.Millisecond, time.Second, task)
	if err != nil {
		t.Fatalf("Expected no error after retries, got %v", err)
	}

	if !result {
		t.Error("Expected true, got false")
	}
}

func TestExecutorWithCallbacks(t *testing.T) {
	var retryCallbacks []int
	var timeoutCallbacks []int

	onRetry := func(attempt int, err error, delay time.Duration) {
		retryCallbacks = append(retryCallbacks, attempt)
	}

	onTimeout := func(attempt int, timeout time.Duration) {
		timeoutCallbacks = append(timeoutCallbacks, attempt)
	}

	attempts := 0
	task := func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("fail")
		}
		return "success", nil
	}

	executor := NewExecutor(
		WithRetryStrategy(NewFixedDelayStrategy(3, 1*time.Millisecond)),
		WithRetryCallback(onRetry),
		WithTimeoutCallback(onTimeout),
	)

	result, err := Execute(executor, context.Background(), task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Value != "success" {
		t.Errorf("Expected 'success', got '%s'", result.Value)
	}

	// Should have 2 retry callbacks (attempts 1 and 2 failed)
	if len(retryCallbacks) != 2 {
		t.Errorf("Expected 2 retry callbacks, got %d", len(retryCallbacks))
	}

	// Should have 3 timeout callbacks (one for each attempt)
	if len(timeoutCallbacks) != 3 {
		t.Errorf("Expected 3 timeout callbacks, got %d", len(timeoutCallbacks))
	}
}

func TestConditionalRetryStrategy(t *testing.T) {
	shouldRetry := func(attempt int, err error) bool {
		// Only retry on specific error
		return err != nil && err.Error() == "retryable"
	}

	strategy := NewConditionalRetryStrategy(3, time.Millisecond, shouldRetry, nil)

	// Should retry on retryable error
	if !strategy.ShouldRetry(1, errors.New("retryable")) {
		t.Error("Should retry on retryable error")
	}

	// Should not retry on different error
	if strategy.ShouldRetry(1, errors.New("non-retryable")) {
		t.Error("Should not retry on non-retryable error")
	}
}

func TestProgressiveTimeoutStrategy(t *testing.T) {
	strategy := NewProgressiveTimeoutStrategy(time.Second, 2.0, 10*time.Second)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 10 * time.Second}, // Capped at max
	}

	for _, tt := range tests {
		result := strategy.GetTimeout(tt.attempt)
		if result != tt.expected {
			t.Errorf("GetTimeout(%d) = %v, want %v", tt.attempt, result, tt.expected)
		}
	}
}

// Example test showing how to use the executor for different types of tasks
func TestExecutorDifferentTaskTypes(t *testing.T) {
	ctx := context.Background()
	executor := NewExecutor()

	// String task
	stringTask := func(ctx context.Context) (string, error) {
		return "hello", nil
	}
	stringResult, err := ExecuteSimple(executor, ctx, stringTask)
	if err != nil || stringResult != "hello" {
		t.Errorf("String task failed: %v, result: %s", err, stringResult)
	}

	// Integer task
	intTask := func(ctx context.Context) (int, error) {
		return 42, nil
	}
	intResult, err := ExecuteSimple(executor, ctx, intTask)
	if err != nil || intResult != 42 {
		t.Errorf("Int task failed: %v, result: %d", err, intResult)
	}

	// Struct task
	type Person struct {
		Name string
		Age  int
	}
	structTask := func(ctx context.Context) (Person, error) {
		return Person{Name: "Alice", Age: 30}, nil
	}
	structResult, err := ExecuteSimple(executor, ctx, structTask)
	if err != nil || structResult.Name != "Alice" {
		t.Errorf("Struct task failed: %v, result: %+v", err, structResult)
	}
}
