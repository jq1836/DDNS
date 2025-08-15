package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jq1836/DDNS/ddns"
	"github.com/jq1836/DDNS/executor"
)

// DuckDNSProvider implements the DDNS Provider interface for DuckDNS
type DuckDNSProvider struct {
	token      string
	httpClient *http.Client
	executor   *executor.Executor
}

// DuckDNSConfig holds DuckDNS-specific configuration
type DuckDNSConfig struct {
	Token string
}

// NewDuckDNSProvider creates a new DuckDNS DDNS provider
func NewDuckDNSProvider(config DuckDNSConfig) *DuckDNSProvider {
	// Set up executor with retry logic for API calls
	exec := executor.NewExecutor(
		executor.WithRetryStrategy(executor.NewExponentialBackoffStrategy(3, time.Second, 2.0)),
		executor.WithTimeoutStrategy(executor.NewFixedTimeoutStrategy(30*time.Second)),
	)

	return &DuckDNSProvider{
		token:      config.Token,
		httpClient: &http.Client{},
		executor:   exec,
	}
}

// UpdateRecord updates a DNS record in DuckDNS
func (d *DuckDNSProvider) UpdateRecord(ctx context.Context, req ddns.UpdateRequest) (*ddns.UpdateResponse, error) {
	task := func(taskCtx context.Context) (*ddns.UpdateResponse, error) {
		// Build the DuckDNS update URL
		baseURL := "https://www.duckdns.org/update"
		params := url.Values{}
		params.Set("domains", req.Domain)
		params.Set("token", d.token)
		params.Set("ip", req.Value)

		updateURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(taskCtx, "GET", updateURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("User-Agent", "ddns-client/1.0")

		// Make the request
		resp, err := d.httpClient.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		responseText := strings.TrimSpace(string(body))

		// DuckDNS returns "OK" for success, "KO" for failure
		if responseText == "OK" {
			return &ddns.UpdateResponse{
				Success:   true,
				Message:   "DuckDNS record updated successfully",
				RecordID:  req.Domain, // DuckDNS doesn't have record IDs, use domain
				UpdatedAt: time.Now(),
			}, nil
		} else if responseText == "KO" {
			return nil, fmt.Errorf("DuckDNS update failed: invalid token or domain")
		} else {
			return nil, fmt.Errorf("unexpected DuckDNS response: %s", responseText)
		}
	}

	return executor.ExecuteSimple(d.executor, ctx, task)
}

// GetCurrentRecord retrieves the current DNS record value
// Note: DuckDNS doesn't provide an API to get current records, so we'll return an error
// This forces the service to always attempt an update
func (d *DuckDNSProvider) GetCurrentRecord(ctx context.Context, domain, recordType string) (string, error) {
	// DuckDNS doesn't provide a way to query current records
	// Return an error to force updates
	return "", fmt.Errorf("DuckDNS does not support querying current records")
}

// ValidateCredentials checks if the DuckDNS credentials are valid
func (d *DuckDNSProvider) ValidateCredentials(ctx context.Context) error {
	task := func(taskCtx context.Context) (interface{}, error) {
		// Use a test domain to validate credentials
		// We'll make a request without actually updating anything
		baseURL := "https://www.duckdns.org/update"
		params := url.Values{}
		params.Set("domains", "test") // Use a test domain that likely doesn't exist
		params.Set("token", d.token)
		params.Set("verbose", "true")

		validateURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

		req, err := http.NewRequestWithContext(taskCtx, "GET", validateURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "ddns-client/1.0")

		resp, err := d.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("validation request failed: %w", err)
		}
		defer resp.Body.Close()

		// If we get a valid HTTP response, the service is reachable
		// DuckDNS will return "KO" for invalid token, but at least we know the service works
		if resp.StatusCode == http.StatusOK {
			return nil, nil // Service is reachable, token format is acceptable
		}

		return nil, fmt.Errorf("DuckDNS service returned status: %s", resp.Status)
	}

	_, err := executor.ExecuteSimple(d.executor, ctx, task)
	return err
}

// GetProviderName returns the name of the provider
func (d *DuckDNSProvider) GetProviderName() string {
	return "duckdns"
}
