package ddns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ddns/executor"
)

// IPResponse represents the response from httpbin.org/ip
type IPResponse struct {
	Origin string `json:"origin"`
}

// getIPFromHTTPBin retrieves the public IP from httpbin.org
func getIPFromHTTPBin(ctx context.Context) (string, error) {
	// Create a task for getting the IP
	ipTask := func(taskCtx context.Context) (string, error) {
		client := &http.Client{}

		req, err := http.NewRequestWithContext(taskCtx, "GET", "https://httpbin.org/ip", nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", "ddns-client/1.0")

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		var ipResp IPResponse
		if err := json.Unmarshal(body, &ipResp); err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}

		if ipResp.Origin == "" {
			return "", fmt.Errorf("no IP address in response")
		}

		return ipResp.Origin, nil
	}

	// Use the executor for retry logic
	exec := executor.NewExecutor(
		executor.WithRetryStrategy(executor.NewExponentialBackoffStrategy(3, time.Second, 2.0)),
		executor.WithTimeoutStrategy(executor.NewFixedTimeoutStrategy(10*time.Second)),
	)

	return executor.ExecuteSimple(exec, ctx, ipTask)
}
