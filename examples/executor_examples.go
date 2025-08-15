package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ddns/executor"
)

// Example 1: Database operation with retries
func databaseOperation(ctx context.Context) (string, error) {
	// Simulate a database operation that might fail
	select {
	case <-time.After(100 * time.Millisecond):
		return "database_result", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Example 2: File operation with timeout
func fileOperation(ctx context.Context) ([]byte, error) {
	// Simulate reading a file
	select {
	case <-time.After(50 * time.Millisecond):
		return []byte("file content"), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Example 3: API call with custom retry logic
func apiCall(ctx context.Context) (map[string]interface{}, error) {
	// Simulate an API call
	result := map[string]interface{}{
		"status": "success",
		"data":   "api response",
	}
	return result, nil
}

// Example 4: Complex business operation
type User struct {
	ID   int
	Name string
}

func createUser(ctx context.Context) (User, error) {
	// Simulate user creation with validation
	return User{ID: 123, Name: "John Doe"}, nil
}

func demonstrateExecutorUsage() {
	ctx := context.Background()

	// Example 1: Database operation with exponential backoff
	fmt.Println("=== Database Operation with Retries ===")
	dbExecutor := executor.NewExecutor(
		executor.WithRetryStrategy(executor.NewExponentialBackoffStrategy(3, time.Second, 2.0)),
		executor.WithTimeoutStrategy(executor.NewFixedTimeoutStrategy(5*time.Second)),
		executor.WithRetryCallback(func(attempt int, err error, delay time.Duration) {
			fmt.Printf("DB operation failed (attempt %d): %v, retrying in %v\n", attempt, err, delay)
		}),
	)

	dbResult, err := executor.ExecuteSimple(dbExecutor, ctx, databaseOperation)
	if err != nil {
		fmt.Printf("Database operation failed: %v\n", err)
	} else {
		fmt.Printf("Database result: %s\n", dbResult)
	}

	// Example 2: File operation with timeout only
	fmt.Println("\n=== File Operation with Timeout ===")
	fileResult, err := executor.ExecuteWithTimeout(ctx, 2*time.Second, fileOperation)
	if err != nil {
		fmt.Printf("File operation failed: %v\n", err)
	} else {
		fmt.Printf("File content: %s\n", string(fileResult))
	}

	// Example 3: API call with linear backoff and retries
	fmt.Println("\n=== API Call with Linear Backoff ===")
	apiResult, err := executor.ExecuteWithRetries(ctx, 2, 500*time.Millisecond, 3*time.Second, apiCall)
	if err != nil {
		fmt.Printf("API call failed: %v\n", err)
	} else {
		fmt.Printf("API result: %+v\n", apiResult)
	}

	// Example 4: Custom conditional retry strategy
	fmt.Println("\n=== User Creation with Custom Retry Logic ===")
	customRetryLogic := func(attempt int, err error) bool {
		// Only retry on specific errors
		if err != nil && err.Error() == "user_exists" {
			return attempt < 3
		}
		return false
	}

	customExecutor := executor.NewExecutor(
		executor.WithRetryStrategy(
			executor.NewConditionalRetryStrategy(3, time.Second, customRetryLogic, nil),
		),
	)

	userResult, err := executor.ExecuteSimple(customExecutor, ctx, createUser)
	if err != nil {
		fmt.Printf("User creation failed: %v\n", err)
	} else {
		fmt.Printf("Created user: %+v\n", userResult)
	}

	// Example 5: Progressive timeout strategy
	fmt.Println("\n=== Operation with Progressive Timeouts ===")
	progressiveExecutor := executor.NewExecutor(
		executor.WithRetryStrategy(executor.NewFixedDelayStrategy(3, 500*time.Millisecond)),
		executor.WithTimeoutStrategy(executor.NewProgressiveTimeoutStrategy(1*time.Second, 1.5, 5*time.Second)),
		executor.WithTimeoutCallback(func(attempt int, timeout time.Duration) {
			fmt.Printf("Attempt %d: timeout set to %v\n", attempt, timeout)
		}),
	)

	progressiveResult, err := executor.ExecuteSimple(progressiveExecutor, ctx, databaseOperation)
	if err != nil {
		fmt.Printf("Progressive operation failed: %v\n", err)
	} else {
		fmt.Printf("Progressive result: %s\n", progressiveResult)
	}
}

// Example usage in real applications
func realWorldExamples() {
	ctx := context.Background()

	// External service call with circuit breaker-like behavior
	serviceCall := func(ctx context.Context) (string, error) {
		// Simulate external service call
		return "service_response", nil
	}

	// Only retry on network errors, not on 4xx responses
	shouldRetryNetworkOnly := func(attempt int, err error) bool {
		if err != nil {
			// In a real implementation, you'd check if it's a network error
			return true
		}
		return false
	}

	networkExecutor := executor.NewExecutor(
		executor.WithRetryStrategy(
			executor.NewConditionalRetryStrategy(3, 2*time.Second, shouldRetryNetworkOnly, nil),
		),
	)

	result, err := executor.ExecuteSimple(networkExecutor, ctx, serviceCall)
	if err != nil {
		log.Printf("Service call failed: %v", err)
	} else {
		log.Printf("Service response: %s", result)
	}
}
